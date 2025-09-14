package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	DYNAMODB_TABLE_STATIONS     string
	ACCESS_CONTROL_ALLOW_ORIGIN string
)

// Response represents the API Gateway response structure
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type HeatingStationAPI struct {
	GeoId      string  `json:"geoId"`
	Name       string  `json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	LastStatus string  `json:"lastStatus"`
}

type ApiResponseData struct {
	Data []HeatingStationAPI `json:"data"`
}

// Sort stations by name
func sortStationsByName(stations []HeatingStationAPI) {
	sort.SliceStable(stations, func(i, j int) bool {
		return stations[i].Name < stations[j].Name
	})
}

// Get stations from DynamoDB table
func getStations(ctx context.Context) ([]HeatingStationAPI, error) {
	// Scan the stations table
	result, err := dbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(DYNAMODB_TABLE_STATIONS),
		Limit:     aws.Int32(10000),
	})
	if err != nil {
		return nil, err
	}

	var stations []scrapper.HeatingStation
	err = attributevalue.UnmarshalListOfMaps(result.Items, &stations)
	if err != nil {
		return nil, err
	}

	// Convert to API format with string geoId
	apiStations := make([]HeatingStationAPI, len(stations))
	for i, station := range stations {
		apiStations[i] = HeatingStationAPI{
			GeoId:      fmt.Sprintf("%d", station.GeoId),
			Name:       station.Name,
			Latitude:   station.Latitude,
			Longitude:  station.Longitude,
			LastStatus: station.LastStatus,
		}
	}

	// Sort stations before returning
	sortStationsByName(apiStations)
	return apiStations, nil
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

	stations, err := getStations(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       `{"message": "Internal server error"}`,
		}, nil
	}

	respData := ApiResponseData{
		Data: stations,
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
