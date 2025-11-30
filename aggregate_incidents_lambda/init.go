package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	dbClient                    *dynamodb.Client
	s3Client                    *s3.Client
	DYNAMODB_TABLE_STATIONS     string
	S3_BUCKET                   string
	ACCESS_CONTROL_ALLOW_ORIGIN string
)

func init() {
	var err error

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("Failed to load AWS SDK config", "error_msg", err.Error())
		panic(err)
	}
	dbClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)

	DYNAMODB_TABLE_STATIONS = os.Getenv("DYNAMODB_TABLE_STATIONS")
	S3_BUCKET = os.Getenv("S3_BUCKET")
	ACCESS_CONTROL_ALLOW_ORIGIN = os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN")
	
	if DYNAMODB_TABLE_STATIONS == "" {
		slog.Error("Required environment variable DYNAMODB_TABLE_STATIONS not set")
		panic("Missing required environment variables")
	}
	if S3_BUCKET == "" {
		slog.Error("Required environment variable S3_BUCKET not set")
		panic("Missing required environment variables")
	}
	if ACCESS_CONTROL_ALLOW_ORIGIN == "" {
		ACCESS_CONTROL_ALLOW_ORIGIN = "*"
	}
}