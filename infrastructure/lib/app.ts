#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "aws-cdk-lib";
import { DatabaseStack } from "./database-stack";
import { EcrStack } from "./ecr-stack";
import { LambdaStack } from "./lambda-stack";
import { ApiStack } from "./api-stack";
import { ScheduleStack } from "./schedule-stack";
import { AlertsStack } from "./alerts-stack";
import { LogsStack } from "./logs-stack";

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

new LogsStack(app, "BucharestTermoficareLogs", {
  env,
  envPrefix,
});

const ecrStack = new EcrStack(app, "BucharestTermoficareEcr", {
  env,
  envPrefix,
});

const lambdaStack = new LambdaStack(app, "BucharestTermoficareLambda", {
  env,
  envPrefix,
  version,
  ecrRepository: ecrStack.repository,
  stationsTable: databaseStack.stationsTable,
  dayCountsTable: databaseStack.dayCountsTable,
  statusHistoryTable: databaseStack.statusHistoryTable,
});

new ScheduleStack(app, "BucharestTermoficareSchedule", {
  env,
  envPrefix,
  lambdaFunction: lambdaStack.lambdaFunction,
  scheduleExpression: "cron(0,30 * * * ? *)", // Every hour at minute 0 (UTC)
});

new ApiStack(app, "BucharestTermoficareApi", {
  env,
  envPrefix,
  version,
  ecrRepository: ecrStack.repository,
  dayCountsTable: databaseStack.dayCountsTable,
  stationsTable: databaseStack.stationsTable,
  statusHistoryTable: databaseStack.statusHistoryTable,
});

if (envPrefix === "prod") {
  if (!alertEmail) {
    throw new Error("alertEmail context required for prod environment");
  }

  new AlertsStack(app, "BucharestTermoficareAlerts", {
    env,
    envPrefix,
    etlLambdaFunction: lambdaStack.lambdaFunction,
    streamProcessorFunction: databaseStack.streamProcessor,
    alertEmail,
  });
}
