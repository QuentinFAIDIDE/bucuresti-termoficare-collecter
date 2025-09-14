package main

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/QuentinFAIDIDE/bucuresti-termoficare-collecter/scrapper"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	dbClient                    *dynamodb.Client
	DYNAMODB_TABLE_DAY_COUNTS   string
	ACCESS_CONTROL_ALLOW_ORIGIN string
)

// Response represents the API Gateway response structure
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type ApiResponseData struct {
	Data []scrapper.StationStatesCount `json:"data"`
}

// Sort counts from most recent to oldest
func sortCountsByDate(counts []scrapper.StationStatesCount) {
	sort.SliceStable(counts, func(i, j int) bool {
		return counts[i].Time > counts[j].Time
	})
}

// Get counts from DynamoDB table
func getCounts(ctx context.Context) ([]scrapper.StationStatesCount, error) {
	// Scan the counts table
	result, err := dbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(DYNAMODB_TABLE_DAY_COUNTS),
		Limit:     aws.Int32(10000),
	})
	if err != nil {
		return nil, err
	}

	var counts []scrapper.StationStatesCount
	err = attributevalue.UnmarshalListOfMaps(result.Items, &counts)
	if err != nil {
		return nil, err
	}

	// Sort counts before returning
	sortCountsByDate(counts)
	return counts, nil
}

// Update handler to return counts
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

	counts, err := getCounts(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       `{"message": "Internal server error"}`,
		}, nil
	}

	respData := ApiResponseData{
		Data: counts,
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
