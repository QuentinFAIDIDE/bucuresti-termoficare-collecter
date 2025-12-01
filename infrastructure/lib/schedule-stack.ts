import * as cdk from "aws-cdk-lib";
import * as events from "aws-cdk-lib/aws-events";
import * as targets from "aws-cdk-lib/aws-events-targets";
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";

interface ScheduleStackProps extends cdk.StackProps {
  envPrefix: string;
  etlLambda: lambda.Function;
  aggregateLambda: lambda.Function;
  scheduleExpression: string;
}

export class ScheduleStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: ScheduleStackProps) {
    super(scope, id, props);

    const etlRule = new events.Rule(this, "ScheduleRule", {
      ruleName: `${props.envPrefix}-termoficare-schedule`,
      schedule: events.Schedule.expression(props.scheduleExpression),
    });

    const aggregateRule = new events.Rule(this, "AggregateScheduleRule", {
      ruleName: `${props.envPrefix}-termoficare-aggregate-schedule`,
      schedule: events.Schedule.expression("cron(0 2 * * ? *)"), // Daily at 2 AM UTC
    });

    etlRule.addTarget(new targets.LambdaFunction(props.etlLambda));
    aggregateRule.addTarget(new targets.LambdaFunction(props.aggregateLambda));
  }
}
