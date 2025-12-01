package main

import (
	"context"
	"encoding/json"
	"fmt"

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

type StationIncidentStatsAPI struct {
	City                        string  `json:"city"`
	Rank                        int     `json:"rank"`
	GeoId                       string  `json:"geoId"`
	LastName                    string  `json:"lastName"`
	Latitude                    float64 `json:"latitude"`
	Longitude                   float64 `json:"longitude"`
	AvgMonthlyIncidentTimeHours float32 `json:"avgMonthlyIncidentTimeHours"`
	AvgIncidentTimeHours        float32 `json:"avgIncidentTimeHours"`
	MaxIncidentTimeHours        float32 `json:"maxIncidentTimeHours"`
}

type ApiResponseData struct {
	Data []StationIncidentStatsAPI `json:"data"`
}

func getStationsStats(ctx context.Context) ([]StationIncidentStatsAPI, error) {

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

	// Convert to API format with string geoId
	apiStats := make([]StationIncidentStatsAPI, len(allStats))
	for i, stat := range allStats {
		apiStats[i] = StationIncidentStatsAPI{
			City:                        stat.City,
			Rank:                        stat.Rank,
			GeoId:                       fmt.Sprintf("%d", stat.GeoId),
			LastName:                    stat.LastName,
			Latitude:                    stat.Latitude,
			Longitude:                   stat.Longitude,
			AvgMonthlyIncidentTimeHours: stat.AvgMonthlyIncidentTimeHours,
			AvgIncidentTimeHours:        stat.AvgIncidentTimeHours,
			MaxIncidentTimeHours:        stat.MaxIncidentTimeHours,
		}
	}

	return apiStats, nil
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
