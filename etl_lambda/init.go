package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/QuentinFAIDIDE/bucuresti-termoficare-collecter/scrapper"
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

	scrapClient, err = scrapper.NewTermoficareScrapper("")
	if err != nil {
		slog.Error("Failed to create TermoficareScrapper", "error_msg", err.Error())
		panic(err)
	}

	DYNAMODB_TABLE_DAY_COUNTS = os.Getenv("DYNAMODB_TABLE_DAY_COUNTS")
	DYNAMODB_TABLE_STATIONS = os.Getenv("DYNAMODB_TABLE_STATIONS")
	DYNAMODB_TABLE_STATUSES = os.Getenv("DYNAMODB_TABLE_STATUSES")
	if DYNAMODB_TABLE_DAY_COUNTS == "" || DYNAMODB_TABLE_STATIONS == "" || DYNAMODB_TABLE_STATUSES == "" {
		slog.Error("Required environment variables DYNAMODB_TABLE_STATIONS and/or DYNAMODB_TABLE_STATUSES and/or DYNAMODB_TABLE_DAY_COUNTS not set")
		panic("Missing required environment variables")
	}
}
