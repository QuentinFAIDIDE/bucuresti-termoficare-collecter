import * as cdk from "aws-cdk-lib";
import * as logs from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";

interface LogsStackProps extends cdk.StackProps {
  envPrefix: string;
}

export class LogsStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: LogsStackProps) {
    super(scope, id, props);

    new logs.LogGroup(this, "WebsiteBackendLogGroup", {
      logGroupName: `${props.envPrefix}-TermoficareWebsiteBackend`,
      retention: logs.RetentionDays.ONE_MONTH,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    new logs.LogGroup(this, "ETLLogGroup", {
      logGroupName: `${props.envPrefix}-TermoficareETL`,
      retention: logs.RetentionDays.ONE_MONTH,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    new logs.LogGroup(this, "ETLBackupStreamProcessorLogGroup", {
      logGroupName: `${props.envPrefix}-TermoficareETLBackupStreamProcessor`,
      retention: logs.RetentionDays.ONE_MONTH,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });
  }
}