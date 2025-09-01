import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ecr from "aws-cdk-lib/aws-ecr";
import { Construct } from "constructs";

interface LambdaStackProps extends cdk.StackProps {
  envPrefix: string;
  version: string;
  stationsTable: dynamodb.Table;
  dayCountsTable: dynamodb.Table;
  statusHistoryTable: dynamodb.Table;
}

export class LambdaStack extends cdk.Stack {
  public readonly lambdaFunction: lambda.Function;
  public readonly ecrRepository: ecr.Repository;

  constructor(scope: Construct, id: string, props: LambdaStackProps) {
    super(scope, id, props);

    this.ecrRepository = new ecr.Repository(this, "LambdaRepository", {
      repositoryName: `${props.envPrefix}-bucuresti-termoficare-lambda`,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    this.lambdaFunction = new lambda.Function(this, "TermoficareLambda", {
      code: lambda.Code.fromEcrImage(this.ecrRepository, {
        tagOrDigest: props.version,
      }),
      handler: lambda.Handler.FROM_IMAGE,
      runtime: lambda.Runtime.FROM_IMAGE,
      timeout: cdk.Duration.minutes(5),
      memorySize: 512,
      environment: {
        DYNAMODB_TABLE_STATIONS: props.stationsTable.tableName,
        DYNAMODB_TABLE_DAY_COUNTS: props.dayCountsTable.tableName,
        DYNAMODB_TABLE_STATUS_HISTORY: props.statusHistoryTable.tableName,
      },
    });

    props.stationsTable.grantReadWriteData(this.lambdaFunction);
    props.dayCountsTable.grantReadWriteData(this.lambdaFunction);
    props.statusHistoryTable.grantReadWriteData(this.lambdaFunction);
  }
}
