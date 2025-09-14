package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func init() {
	var err error

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("Failed to load AWS SDK config", "error_msg", err.Error())
		panic(err)
	}
	dbClient = dynamodb.NewFromConfig(cfg)

	DYNAMODB_TABLE_DAY_COUNTS = os.Getenv("DYNAMODB_TABLE_DAY_COUNTS")
	ACCESS_CONTROL_ALLOW_ORIGIN = os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN")
	if DYNAMODB_TABLE_DAY_COUNTS == "" {
		slog.Error("Required environment variable DYNAMODB_TABLE_DAY_COUNTS not set")
		panic("Missing required environment variables")
	}
	if ACCESS_CONTROL_ALLOW_ORIGIN == "" {
		ACCESS_CONTROL_ALLOW_ORIGIN = "*"
	}
}