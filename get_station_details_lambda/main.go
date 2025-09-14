package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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
	DYNAMODB_TABLE_STATUS_HISTORY string
	ACCESS_CONTROL_ALLOW_ORIGIN   string
)

// Response represents the API Gateway response structure
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type ApiResponseData struct {
	Data []scrapper.HeatingStationStatus `json:"data"`
}

// Get station statuses from DynamoDB table by geoId
func getStationStatuses(ctx context.Context, geoId int64) ([]scrapper.HeatingStationStatus, error) {
	// Query the status history table by geoId
	result, err := dbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(DYNAMODB_TABLE_STATUS_HISTORY),
		KeyConditionExpression: aws.String("GeoId = :geoId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":geoId": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", geoId)},
		},
		ScanIndexForward: aws.Bool(false), // Descending order by sort key (Timestamp)
		Limit:           aws.Int32(5000),
	})
	if err != nil {
		return nil, err
	}

	var statuses []scrapper.HeatingStationStatus
	err = attributevalue.UnmarshalListOfMaps(result.Items, &statuses)
	if err != nil {
		return nil, err
	}

	return statuses, nil
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

	// Get geoId from query parameters
	geoIdStr := request.QueryStringParameters["geoId"]
	if geoIdStr == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       `{"message": "Missing geoId query parameter"}`,
		}, nil
	}

	geoId, err := strconv.ParseInt(geoIdStr, 10, 64)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       `{"message": "Invalid geoId parameter"}`,
		}, nil
	}

	statuses, err := getStationStatuses(ctx, geoId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       `{"message": "Internal server error"}`,
		}, nil
	}

	respData := ApiResponseData{
		Data: statuses,
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
