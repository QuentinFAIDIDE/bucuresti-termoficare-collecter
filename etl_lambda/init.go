package main

import (
	"context"
	"log/slog"

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
}
