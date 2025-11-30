package main

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/QuentinFAIDIDE/bucuresti-termoficare-collecter/scrapper"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var ErrDayBackupNotFound = errors.New("day backup not found")

// Handler processes EventBridge schedule rule events
func Handler(ctx context.Context, event events.CloudWatchEvent) error {

	const (
		maxMissingBackupDays     = 3
		lastDayWithDataInBackups = "2025-11-03"
	)

	slog.Info("Starting rank stations processing")

	currentDayTimestamp := time.Now()
	cutoffTimestamp := time.Now().AddDate(-1, 0, 0)
	dataset := make([]scrapper.HeatingStationStatus, 0, 24*2*1000*365)
	missingBackupDays := 0
	lastDayWithData := ""
	currentDateStr := ""

	for currentDayTimestamp.After(cutoffTimestamp) {

		currentDateStr = currentDayTimestamp.Format("2006-01-02")

		if missingBackupDays >= maxMissingBackupDays {
			slog.Error(
				"Reached maximimum number of days without data, aborting data fetch...",
				"CurrentDatasetSize", len(dataset),
				"maxDaysWithoutData", maxMissingBackupDays,
				"currentDateStr", currentDateStr,
			)
			break
		}

		err := appendDayBackupToDataset(ctx, dataset, currentDateStr)
		if err != nil && err != ErrDayBackupNotFound {
			return err
		}

		if err == ErrDayBackupNotFound {
			slog.Warn("Day backup not found", "date", currentDateStr)
			missingBackupDays++
		} else {
			lastDayWithData = currentDateStr
			missingBackupDays = 0
		}

		currentDayTimestamp = currentDayTimestamp.AddDate(0, 0, -1)
	}

	// if we didnt get a year of data from the daily backups
	if currentDayTimestamp.After(cutoffTimestamp) {
		// if its because we reached the period before the automated backups
		if lastDayWithData == lastDayWithDataInBackups {
			// we load the db backup file for the dates before
			err := loadDDBBackup(ctx, dataset)
			if err != nil {
				return err
			}
			// if we miss data and the backup was not next, it means we have a gap in the backups
		} else {
			slog.Error(
				"A gap was detected in the kinesis stream backups that is larger than max allowed. We need a gap friendly algorithm",
				"lastDayBeforeGap", lastDayWithData,
			)
			return errors.New("a gap in the stations status history was found, please implement a gap-friendly algorithm")
		}
	}

	filteredDataset := scrapper.FilterDataset(dataset, cutoffTimestamp)

	stationsIncidentStats := scrapper.ComputeIncidentStatistics(filteredDataset)

	err := writeStationsIncidentStats(ctx, stationsIncidentStats)
	if err != nil {
		return err
	}

	return nil
}

type BackupRecord struct {
	Timestamp float64 `json:"timestamp"`
	Item      struct {
		IncidentText     string  `json:"IncidentText"`
		Status           string  `json:"Status"`
		IncidentType     string  `json:"IncidentType"`
		GeoId            int64   `json:"GeoId"`
		EstimatedFixDate int64   `json:"EstimatedFixDate"`
		Latitude         float64 `json:"Latitude"`
		Longitude        float64 `json:"Longitude"`
		Timestamp        int64   `json:"Timestamp"`
		Name             string  `json:"Name"`
	} `json:"item"`
}

func appendDayBackupToDataset(ctx context.Context, dataset []scrapper.HeatingStationStatus, dateStr string) error {
	// Check if folder exists
	result, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(S3_BUCKET),
		Prefix:  aws.String(dateStr + "/"),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("failed to check if folder %s exists: %w", dateStr, err)
	}
	if len(result.Contents) == 0 {
		return ErrDayBackupNotFound
	}

	// List all .json.gz files in the folder
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(S3_BUCKET),
		Prefix: aws.String(dateStr + "/"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects in folder %s: %w", dateStr, err)
		}

		for _, obj := range page.Contents {
			if !strings.HasSuffix(*obj.Key, ".json.gz") {
				slog.Warn("a file in the s3 folder was not of .json.gz file", "folderName", dateStr)
				continue
			}

			// Download file
			getResult, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(S3_BUCKET),
				Key:    obj.Key,
			})
			if err != nil {
				return fmt.Errorf("failed to download file %s: %w", *obj.Key, err)
			}
			defer getResult.Body.Close()

			// Extract gzip
			gzReader, err := gzip.NewReader(getResult.Body)
			if err != nil {
				return fmt.Errorf("failed to create gzip reader for %s: %w", *obj.Key, err)
			}
			defer gzReader.Close()

			// Parse entire JSON array
			var records []BackupRecord
			decoder := json.NewDecoder(gzReader)
			if err := decoder.Decode(&records); err != nil {
				return fmt.Errorf("failed to parse JSON from %s: %w", *obj.Key, err)
			}

			// Convert all records to HeatingStationStatus
			for _, record := range records {
				status := scrapper.HeatingStationStatus{
					GeoId:            record.Item.GeoId,
					Name:             record.Item.Name,
					Latitude:         record.Item.Latitude,
					Longitude:        record.Item.Longitude,
					Status:           record.Item.Status,
					IncidentText:     record.Item.IncidentText,
					IncidentType:     record.Item.IncidentType,
					FetchTime:        record.Item.Timestamp,
					EstimatedFixDate: record.Item.EstimatedFixDate,
				}
				dataset = append(dataset, status)
			}
		}
	}

	return nil
}

func loadDDBBackup(ctx context.Context, dataset []scrapper.HeatingStationStatus) error {
	// Fetch dynamodb_backup.csv.gz from S3 root
	getResult, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(S3_BUCKET),
		Key:    aws.String("dynamodb_backup.csv.gz"),
	})
	if err != nil {
		return fmt.Errorf("failed to download dynamodb_backup.csv.gz: %w", err)
	}
	defer getResult.Body.Close()

	// Extract gzip
	gzReader, err := gzip.NewReader(getResult.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader for CSV backup: %w", err)
	}
	defer gzReader.Close()

	// Parse CSV
	csvReader := csv.NewReader(gzReader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV backup: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	// Map column names to indices
	header := records[0]
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	// Parse each row
	for _, row := range records[1:] {
		geoId, err := strconv.ParseInt(row[columnMap["GeoId"]], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse GeoId: %w", err)
		}
		latitude, err := strconv.ParseFloat(row[columnMap["Latitude"]], 64)
		if err != nil {
			return fmt.Errorf("failed to parse Latitude: %w", err)
		}
		longitude, err := strconv.ParseFloat(row[columnMap["Longitude"]], 64)
		if err != nil {
			return fmt.Errorf("failed to parse Longitude: %w", err)
		}
		timestamp, err := strconv.ParseInt(row[columnMap["Timestamp"]], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse Timestamp: %w", err)
		}
		estimatedFixDate, err := strconv.ParseInt(row[columnMap["EstimatedFixDate"]], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse EstimatedFixDate: %w", err)
		}

		status := scrapper.HeatingStationStatus{
			GeoId:            geoId,
			Name:             row[columnMap["Name"]],
			Latitude:         latitude,
			Longitude:        longitude,
			Status:           row[columnMap["Status"]],
			IncidentText:     row[columnMap["IncidentText"]],
			IncidentType:     row[columnMap["IncidentType"]],
			FetchTime:        timestamp,
			EstimatedFixDate: estimatedFixDate,
		}
		dataset = append(dataset, status)
	}

	return nil
}

func writeStationsIncidentStats(ctx context.Context, stats []scrapper.StationIncidentStatsDbRow) error {

	for _, stat := range stats {
		item, err := attributevalue.MarshalMap(stat)
		if err != nil {
			return fmt.Errorf("failed to marshal station stats: %w", err)
		}

		_, err = dbClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("StationIncidentsStats"),
			Item:      item,
		})
		if err != nil {
			return fmt.Errorf("failed to write station stats to DynamoDB: %w", err)
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
