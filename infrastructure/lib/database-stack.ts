import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import { Construct } from "constructs";

interface DatabaseStackProps extends cdk.StackProps {
  envPrefix: string;
}

export class DatabaseStack extends cdk.Stack {
  public readonly stationsTable: dynamodb.Table;
  public readonly dayCountsTable: dynamodb.Table;
  public readonly statusHistoryTable: dynamodb.Table;

  constructor(scope: Construct, id: string, props: DatabaseStackProps) {
    super(scope, id, props);

    this.stationsTable = new dynamodb.Table(this, "StationsTable", {
      tableName: `${props.envPrefix}-stations`,
      partitionKey: { name: "GeoId", type: dynamodb.AttributeType.NUMBER },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // TODO: break this table into partitions by years
    this.dayCountsTable = new dynamodb.Table(this, "DayCountsTable", {
      tableName: `${props.envPrefix}-day-counts`,
      partitionKey: { name: "Timestamp", type: dynamodb.AttributeType.NUMBER },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    this.statusHistoryTable = new dynamodb.Table(this, "StatusHistoryTable", {
      tableName: `${props.envPrefix}-status-history`,
      partitionKey: { name: "GeoId", type: dynamodb.AttributeType.NUMBER },
      sortKey: { name: "Timestamp", type: dynamodb.AttributeType.NUMBER },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });
  }
}
