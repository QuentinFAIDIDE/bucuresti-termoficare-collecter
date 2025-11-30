import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as logs from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";

interface LambdaStackProps extends cdk.StackProps {
  envPrefix: string;
  version: string;
  ecrRepository: cdk.aws_ecr.IRepository;
  stationsTable: dynamodb.Table;
  dayCountsTable: dynamodb.Table;
  statusHistoryTable: dynamodb.Table;
  stationsIncidentsStatsTable: dynamodb.Table;
  backupBucket: s3.Bucket;
}

export class LambdaStack extends cdk.Stack {
  public readonly etlLambda: lambda.Function;
  public readonly aggregateLambda: lambda.Function;

  constructor(scope: Construct, id: string, props: LambdaStackProps) {
    super(scope, id, props);

    const etlLogGroup = logs.LogGroup.fromLogGroupName(
      this,
      "ETLLogGroup",
      `${props.envPrefix}-TermoficareETL`
    );

    const aggregatorLogGroup = logs.LogGroup.fromLogGroupName(
      this,
      "AggregateLogGroup",
      `${props.envPrefix}-TermoficareAggregator`
    );

    this.etlLambda = new lambda.Function(this, "TermoficareLambda", {
      code: lambda.Code.fromEcrImage(props.ecrRepository, {
        tagOrDigest: `etl-${props.version}`,
      }),
      handler: lambda.Handler.FROM_IMAGE,
      runtime: lambda.Runtime.FROM_IMAGE,
      timeout: cdk.Duration.minutes(5),
      memorySize: 512,
      logGroup: etlLogGroup,
      environment: {
        DYNAMODB_TABLE_STATIONS: props.stationsTable.tableName,
        DYNAMODB_TABLE_DAY_COUNTS: props.dayCountsTable.tableName,
        DYNAMODB_TABLE_STATUSES: props.statusHistoryTable.tableName,
      },
    });
    props.stationsTable.grantReadWriteData(this.etlLambda);
    props.dayCountsTable.grantReadWriteData(this.etlLambda);
    props.statusHistoryTable.grantReadWriteData(this.etlLambda);

    this.aggregateLambda = new lambda.Function(this, "AggregateLambda", {
      code: lambda.Code.fromEcrImage(props.ecrRepository, {
        tagOrDigest: `aggregate-${props.version}`,
      }),
      handler: lambda.Handler.FROM_IMAGE,
      runtime: lambda.Runtime.FROM_IMAGE,
      timeout: cdk.Duration.minutes(5),
      memorySize: 512,
      logGroup: aggregatorLogGroup,
      environment: {
        DYNAMODB_TABLE_STATIONS: props.stationsIncidentsStatsTable.tableName,
        S3_BUCKET: props.backupBucket.bucketName,
      },
    });
    props.stationsIncidentsStatsTable.grantWriteData(this.aggregateLambda);
    props.backupBucket.grantRead(this.aggregateLambda);
  }
}
