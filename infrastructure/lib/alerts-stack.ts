import * as cdk from "aws-cdk-lib";
import * as cloudwatch from "aws-cdk-lib/aws-cloudwatch";
import * as cloudwatchActions from "aws-cdk-lib/aws-cloudwatch-actions";
import * as sns from "aws-cdk-lib/aws-sns";
import * as snsSubscriptions from "aws-cdk-lib/aws-sns-subscriptions";
import * as lambda from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";

interface AlertsStackProps extends cdk.StackProps {
  envPrefix: string;
  lambdaFunction: lambda.Function;
  alertEmail: string;
}

export class AlertsStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: AlertsStackProps) {
    super(scope, id, props);

    const topic = new sns.Topic(this, "AlertTopic", {
      topicName: `${props.envPrefix}-termoficare-alerts`,
    });

    topic.addSubscription(
      new snsSubscriptions.EmailSubscription(props.alertEmail)
    );

    const errorAlarm = new cloudwatch.Alarm(this, "LambdaErrorAlarm", {
      alarmName: `${props.envPrefix}-termoficare-lambda-errors`,
      metric: props.lambdaFunction.metricErrors({
        period: cdk.Duration.minutes(5),
      }),
      threshold: 1,
      evaluationPeriods: 1,
      datapointsToAlarm: 1,
      treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
    });
    errorAlarm.addAlarmAction(new cloudwatchActions.SnsAction(topic));

    const missingExecutionAlarm = new cloudwatch.Alarm(
      this,
      "LambdaMissingExecutionAlarm",
      {
        alarmName: `${props.envPrefix}-termoficare-lambda-missing-execution`,
        metric: props.lambdaFunction.metricInvocations({
          period: cdk.Duration.hours(12), // Slightly longer than 6h schedule
        }),
        threshold: 1,
        comparisonOperator: cloudwatch.ComparisonOperator.LESS_THAN_THRESHOLD,
        evaluationPeriods: 1,
        treatMissingData: cloudwatch.TreatMissingData.BREACHING,
      }
    );
    missingExecutionAlarm.addAlarmAction(
      new cloudwatchActions.SnsAction(topic)
    );
  }
}
