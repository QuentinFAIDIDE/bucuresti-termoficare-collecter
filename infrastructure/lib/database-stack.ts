import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as lambdaEventSources from "aws-cdk-lib/aws-lambda-event-sources";
import * as logs from "aws-cdk-lib/aws-logs";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as iam from "aws-cdk-lib/aws-iam";
import { Construct } from "constructs";

interface DatabaseStackProps extends cdk.StackProps {
  envPrefix: string;
}

export class DatabaseStack extends cdk.Stack {
  public readonly stationsTable: dynamodb.Table;
  public readonly dayCountsTable: dynamodb.Table;
  public readonly statusHistoryTable: dynamodb.Table;
  public readonly backupBucket: s3.Bucket;
  public readonly streamProcessor: lambda.Function;

  constructor(scope: Construct, id: string, props: DatabaseStackProps) {
    super(scope, id, props);

    this.stationsTable = new dynamodb.Table(this, "StationsTable", {
      tableName: `${props.envPrefix}-stations`,
      partitionKey: { name: "GeoId", type: dynamodb.AttributeType.NUMBER },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      pointInTimeRecoverySpecification: {
        pointInTimeRecoveryEnabled: true,
      },
    });

    // TODO: break this table into partitions by years
    this.dayCountsTable = new dynamodb.Table(this, "DayCountsTable", {
      tableName: `${props.envPrefix}-day-counts`,
      partitionKey: { name: "Timestamp", type: dynamodb.AttributeType.NUMBER },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      pointInTimeRecoverySpecification: {
        pointInTimeRecoveryEnabled: true,
      },
    });

    this.statusHistoryTable = new dynamodb.Table(this, "StatusHistoryTable", {
      tableName: `${props.envPrefix}-status-history`,
      partitionKey: { name: "GeoId", type: dynamodb.AttributeType.NUMBER },
      sortKey: { name: "Timestamp", type: dynamodb.AttributeType.NUMBER },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      stream: dynamodb.StreamViewType.NEW_IMAGE,
      pointInTimeRecoverySpecification: {
        pointInTimeRecoveryEnabled: true,
      },
    });

    // S3 bucket for backups
    this.backupBucket = new s3.Bucket(this, "BackupBucket", {
      bucketName: `${props.envPrefix}-termoficare-backups`,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });

    const streamLogGroup = logs.LogGroup.fromLogGroupName(
      this,
      "StreamProcessorLogGroup",
      `${props.envPrefix}-TermoficareETLBackupStreamProcessor`
    );

    // Lambda to process DynamoDB streams and write to S3
    this.streamProcessor = new lambda.Function(this, "StreamProcessor", {
      runtime: lambda.Runtime.PYTHON_3_11,
      handler: "stations_ddb_stream_backup.lambda_handler",
      code: lambda.Code.fromAsset("resources"),
      logGroup: streamLogGroup,
      environment: {
        BACKUP_BUCKET: this.backupBucket.bucketName,
      },
      timeout: cdk.Duration.seconds(30),
    });

    // Grant S3 permissions
    this.backupBucket.grantWrite(this.streamProcessor);

    // Connect DynamoDB stream to Lambda
    this.streamProcessor.addEventSource(
      new lambdaEventSources.DynamoEventSource(this.statusHistoryTable, {
        startingPosition: lambda.StartingPosition.LATEST,
        batchSize: 1000,
        maxBatchingWindow: cdk.Duration.seconds(300),
      })
    );
  }
}
