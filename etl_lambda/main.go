package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/QuentinFAIDIDE/bucuresti-termoficare-collecter/scrapper"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	errMissingEnvar                  = errors.New("missing environement variable")
	scrapClient                      *scrapper.TermoficareScrapper
	dbClient                         *dynamodb.Client
	DYNAMODB_TABLE_RANKED_DAY_COUNTS string
	DYNAMODB_TABLE_WORDS_DAY_COUNTS  string
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

	greenCount, yellowCount, redCount, err := scrapClient.GetStatesCounts()
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

	// TODO: push counts to a dynamodb table

	// TODO: push all stations

	// TODO: push all statuses

	// This is an example of pushing some data to dynamodb
	/*
		perSrWordItem := models.DayCountsPerSubredditWordItem{}.WithRedditWordCount(ev.Subreddit, wc, yesterday)
		av2, err := attributevalue.MarshalMap(perSrWordItem)
		if err != nil {
			slog.Error("Unable to Marshal per subreddit word count item", "error_msg", err.Error())
			return err
		}
		putItemInput2 := &dynamodb.PutItemInput{
			TableName: aws.String(DYNAMODB_TABLE_WORDS_DAY_COUNTS),
			Item:      av2,
		}
		_, err = dbClient.PutItem(ctx, putItemInput2)
		if err != nil {
			slog.Error("Unable to write per subreddit word count item", "error_msg", err.Error())
			return err
		}
	*/

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
