import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as logs from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";

interface LambdaStackProps extends cdk.StackProps {
  envPrefix: string;
  version: string;
  ecrRepository: cdk.aws_ecr.IRepository;
  stationsTable: dynamodb.Table;
  dayCountsTable: dynamodb.Table;
  statusHistoryTable: dynamodb.Table;
}

export class LambdaStack extends cdk.Stack {
  public readonly lambdaFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: LambdaStackProps) {
    super(scope, id, props);

    const logGroup = logs.LogGroup.fromLogGroupName(
      this,
      "ETLLogGroup",
      `${props.envPrefix}-TermoficareETL`
    );

    this.lambdaFunction = new lambda.Function(this, "TermoficareLambda", {
      code: lambda.Code.fromEcrImage(props.ecrRepository, {
        tagOrDigest: `etl-${props.version}`,
      }),
      handler: lambda.Handler.FROM_IMAGE,
      runtime: lambda.Runtime.FROM_IMAGE,
      timeout: cdk.Duration.minutes(5),
      memorySize: 512,
      logGroup,
      environment: {
        DYNAMODB_TABLE_STATIONS: props.stationsTable.tableName,
        DYNAMODB_TABLE_DAY_COUNTS: props.dayCountsTable.tableName,
        DYNAMODB_TABLE_STATUSES: props.statusHistoryTable.tableName,
      },
    });

    props.stationsTable.grantReadWriteData(this.lambdaFunction);
    props.dayCountsTable.grantReadWriteData(this.lambdaFunction);
    props.statusHistoryTable.grantReadWriteData(this.lambdaFunction);
  }
}
