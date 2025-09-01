#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "aws-cdk-lib";
import { DatabaseStack } from "./database-stack";
import { LambdaStack } from "./lambda-stack";
import { ScheduleStack } from "./schedule-stack";
import { AlertsStack } from "./alerts-stack";

const app = new cdk.App();
const envPrefix = app.node.tryGetContext("envPrefix") || "dev";
const version = app.node.tryGetContext("version") || "latest";
const alertEmail = app.node.tryGetContext("alertEmail");

const env = {
  account: process.env.CDK_DEFAULT_ACCOUNT,
  region: process.env.CDK_DEFAULT_REGION || "eu-south-2",
};

const databaseStack = new DatabaseStack(app, "BucharestTermoficareDatabase", {
  env,
  envPrefix,
});

const lambdaStack = new LambdaStack(app, "BucharestTermoficareLambda", {
  env,
  envPrefix,
  version,
  stationsTable: databaseStack.stationsTable,
  dayCountsTable: databaseStack.dayCountsTable,
  statusHistoryTable: databaseStack.statusHistoryTable,
});

new ScheduleStack(app, "BucharestTermoficareSchedule", {
  env,
  envPrefix,
  lambdaFunction: lambdaStack.lambdaFunction,
  scheduleExpression: "cron(0 3,9,15,20 * * ? *)", // UTC times
});

if (envPrefix === "prod") {
  if (!alertEmail) {
    throw new Error("alertEmail context required for prod environment");
  }

  new AlertsStack(app, "BucharestTermoficareAlerts", {
    env,
    envPrefix,
    lambdaFunction: lambdaStack.lambdaFunction,
    alertEmail,
  });
}
