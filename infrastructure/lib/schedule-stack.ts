import * as cdk from 'aws-cdk-lib';
import * as events from 'aws-cdk-lib/aws-events';
import * as targets from 'aws-cdk-lib/aws-events-targets';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import { Construct } from 'constructs';

interface ScheduleStackProps extends cdk.StackProps {
  envPrefix: string;
  lambdaFunction: lambda.Function;
  scheduleExpression: string;
}

export class ScheduleStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: ScheduleStackProps) {
    super(scope, id, props);

    const rule = new events.Rule(this, 'ScheduleRule', {
      ruleName: `${props.envPrefix}-termoficare-schedule`,
      schedule: events.Schedule.expression(props.scheduleExpression),
    });

    rule.addTarget(new targets.LambdaFunction(props.lambdaFunction));
  }
}