package main

import (
	"context"
	"encoding/json"

	"github.com/QuentinFAIDIDE/bucuresti-termoficare-collecter/scrapper"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dbClient                      *dynamodb.Client
	DYNAMODB_TABLE_STATIONS_STATS string
	ACCESS_CONTROL_ALLOW_ORIGIN   string
)

type ApiResponseData struct {
	Data []scrapper.StationIncidentStatsDbRow `json:"data"`
}

func getStationsStats(ctx context.Context) ([]scrapper.StationIncidentStatsDbRow, error) {

	var lastKey map[string]types.AttributeValue
	allStats := make([]scrapper.StationIncidentStatsDbRow, 0, 1024)

	for {
		input := &dynamodb.QueryInput{
			TableName:              aws.String(DYNAMODB_TABLE_STATIONS_STATS),
			KeyConditionExpression: aws.String("City = :city"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":city": &types.AttributeValueMemberS{Value: "Bucharest"},
			},
		}

		if lastKey != nil {
			input.ExclusiveStartKey = lastKey
		}

		result, err := dbClient.Query(ctx, input)
		if err != nil {
			return nil, err
		}

		var pageStats []scrapper.StationIncidentStatsDbRow
		err = attributevalue.UnmarshalListOfMaps(result.Items, &pageStats)
		if err != nil {
			return nil, err
		}

		allStats = append(allStats, pageStats...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	return allStats, nil
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Access-Control-Allow-Origin":  ACCESS_CONTROL_ALLOW_ORIGIN,
		"Access-Control-Allow-Methods": "GET,OPTIONS",
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
	}

	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	if request.HTTPMethod != "GET" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       `{"message": "Invalid request method"}`,
		}, nil
	}

	stats, err := getStationsStats(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       `{"message": "Internal server error"}`,
		}, nil
	}

	respData := ApiResponseData{
		Data: stats,
	}

	jsonData, err := json.Marshal(respData)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       `{"message": "Error marshaling response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(jsonData),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
