import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as ecr from "aws-cdk-lib/aws-ecr";
import { Construct } from "constructs";

interface ApiStackProps extends cdk.StackProps {
  envPrefix: string;
  version: string;
  ecrRepository: ecr.Repository;
  dayCountsTable: dynamodb.Table;
  stationsTable: dynamodb.Table;
  statusHistoryTable: dynamodb.Table;
}

export class ApiStack extends cdk.Stack {
  public readonly apiGateway: apigateway.RestApi;
  public readonly getCountsLambda: lambda.Function;
  public readonly getStationsLambda: lambda.Function;
  public readonly getStationDetailsLambda: lambda.Function;

  constructor(scope: Construct, id: string, props: ApiStackProps) {
    super(scope, id, props);

    this.getCountsLambda = new lambda.Function(this, "GetCountsLambda", {
      code: lambda.Code.fromEcrImage(props.ecrRepository, {
        tagOrDigest: `api-getcounts-${props.version}`,
      }),
      handler: lambda.Handler.FROM_IMAGE,
      runtime: lambda.Runtime.FROM_IMAGE,
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      environment: {
        DYNAMODB_TABLE_DAY_COUNTS: props.dayCountsTable.tableName,
        ACCESS_CONTROL_ALLOW_ORIGIN: "*",
      },
    });

    props.dayCountsTable.grantReadData(this.getCountsLambda);

    this.getStationsLambda = new lambda.Function(this, "GetStationsLambda", {
      code: lambda.Code.fromEcrImage(props.ecrRepository, {
        tagOrDigest: `api-getstations-${props.version}`,
      }),
      handler: lambda.Handler.FROM_IMAGE,
      runtime: lambda.Runtime.FROM_IMAGE,
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      environment: {
        DYNAMODB_TABLE_STATIONS: props.stationsTable.tableName,
        ACCESS_CONTROL_ALLOW_ORIGIN: "*",
      },
    });

    props.stationsTable.grantReadData(this.getStationsLambda);

    this.getStationDetailsLambda = new lambda.Function(this, "GetStationDetailsLambda", {
      code: lambda.Code.fromEcrImage(props.ecrRepository, {
        tagOrDigest: `api-getstationdetails-${props.version}`,
      }),
      handler: lambda.Handler.FROM_IMAGE,
      runtime: lambda.Runtime.FROM_IMAGE,
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      environment: {
        DYNAMODB_TABLE_STATUS_HISTORY: props.statusHistoryTable.tableName,
        ACCESS_CONTROL_ALLOW_ORIGIN: "*",
      },
    });

    props.statusHistoryTable.grantReadData(this.getStationDetailsLambda);

    this.apiGateway = new apigateway.RestApi(this, "TermoficareApi", {
      restApiName: `${props.envPrefix}-termoficare-api`,
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: [
          "Content-Type",
          "X-Amz-Date",
          "Authorization",
          "X-Api-Key",
        ],
      },
    });

    const countsResource = this.apiGateway.root.addResource("counts");
    countsResource.addMethod(
      "GET",
      new apigateway.LambdaIntegration(this.getCountsLambda)
    );

    const stationsResource = this.apiGateway.root.addResource("stations");
    stationsResource.addMethod(
      "GET",
      new apigateway.LambdaIntegration(this.getStationsLambda)
    );

    const stationDetailsResource = this.apiGateway.root.addResource("station-details");
    stationDetailsResource.addMethod(
      "GET",
      new apigateway.LambdaIntegration(this.getStationDetailsLambda)
    );

    new cdk.CfnOutput(this, 'ApiUrl', {
      value: this.apiGateway.url,
      description: 'API Gateway URL',
    });

    new cdk.CfnOutput(this, 'CountsEndpoint', {
      value: `${this.apiGateway.url}counts`,
      description: 'Counts API endpoint',
    });

    new cdk.CfnOutput(this, 'StationsEndpoint', {
      value: `${this.apiGateway.url}stations`,
      description: 'Stations API endpoint',
    });

    new cdk.CfnOutput(this, 'StationDetailsEndpoint', {
      value: `${this.apiGateway.url}station-details?geoId=123`,
      description: 'Station details API endpoint (with geoId parameter)',
    });
  }
}
