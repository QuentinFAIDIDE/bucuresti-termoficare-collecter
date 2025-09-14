package main

import (
	"context"
	"log/slog"

	"github.com/QuentinFAIDIDE/bucuresti-termoficare-collecter/scrapper"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	scrapClient               *scrapper.TermoficareScrapper
	dbClient                  *dynamodb.Client
	DYNAMODB_TABLE_DAY_COUNTS string
	DYNAMODB_TABLE_STATIONS   string
	DYNAMODB_TABLE_STATUSES   string
)

func HandleRequest(ctx context.Context, ev events.CloudWatchEvent) error {

	slog.Info("Lambda handling event",
		"id", ev.ID,
		"source", ev.Source,
		"account", ev.AccountID,
		"region", ev.Region,
		"version", ev.Version,
		"time", ev.Time,
		"detail_type", ev.DetailType,
		"detail", ev.Detail,
		"resources", ev.Resources,
	)

	err := scrapClient.PullData()
	if err != nil {
		slog.Error("Unable to pull data", "error_msg", err.Error())
		return err
	}

	counts, err := scrapClient.GetStatesCounts()
	if err != nil {
		slog.Error("Unable to get states counts", "error_msg", err.Error())
		return err
	}

	stations, err := scrapClient.GetHeatingStations()
	if err != nil {
		slog.Error("Unable to get heating stations", "error_msg", err.Error())
		return err
	}

	statuses, err := scrapClient.GetHeatingStationsStatuses()
	if err != nil {
		slog.Error("Unable to get heating stations statuses", "error_msg", err.Error())
		return err
	}

	countsDbItem, err := attributevalue.MarshalMap(counts)
	if err != nil {
		slog.Error("Unable to Marshal counts item", "error_msg", err.Error())
		return err
	}
	putCountDbItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(DYNAMODB_TABLE_DAY_COUNTS),
		Item:      countsDbItem,
	}
	// TODO: break this table into partitions by years
	_, err = dbClient.PutItem(ctx, putCountDbItemInput)
	if err != nil {
		slog.Error("Unable to write day count items", "error_msg", err.Error())
		return err
	}

	for _, station := range stations {
		stationDbItem, err := attributevalue.MarshalMap(station)
		if err != nil {
			slog.Error("Unable to Marshal station item", "error_msg", err.Error())
			return err
		}
		putStationDbItemInput := &dynamodb.PutItemInput{
			TableName: aws.String(DYNAMODB_TABLE_STATIONS),
			Item:      stationDbItem,
		}
		_, err = dbClient.PutItem(ctx, putStationDbItemInput)
		if err != nil {
			slog.Error("Unable to write station item", "error_msg", err.Error())
			return err
		}
	}

	for _, status := range statuses {
		statusDbItem, err := attributevalue.MarshalMap(status)
		if err != nil {
			slog.Error("Unable to Marshal status item", "error_msg", err.Error())
			return err
		}
		putStatusDbItemInput := &dynamodb.PutItemInput{
			TableName: aws.String(DYNAMODB_TABLE_STATUSES),
			Item:      statusDbItem,
		}
		_, err = dbClient.PutItem(ctx, putStatusDbItemInput)
		if err != nil {
			slog.Error("Unable to write status item", "error_msg", err.Error())
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
