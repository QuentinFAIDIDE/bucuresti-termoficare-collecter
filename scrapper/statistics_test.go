package scrapper

import (
	"testing"
	"time"
)

func TestFilterDataset(t *testing.T) {
	cutoff := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		dataset  []HeatingStationStatus
		expected int
	}{
		{
			name:     "empty dataset",
			dataset:  []HeatingStationStatus{},
			expected: 0,
		},
		{
			name: "all items after cutoff",
			dataset: []HeatingStationStatus{
				{FetchTime: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC).Unix()},
				{FetchTime: time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC).Unix()},
			},
			expected: 2,
		},
		{
			name: "all items before cutoff",
			dataset: []HeatingStationStatus{
				{FetchTime: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC).Unix()},
				{FetchTime: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC).Unix()},
			},
			expected: 0,
		},
		{
			name: "mixed items",
			dataset: []HeatingStationStatus{
				{FetchTime: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC).Unix()},
				{FetchTime: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC).Unix()},
				{FetchTime: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC).Unix()},
				{FetchTime: time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC).Unix()},
				{FetchTime: time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC).Unix()},
			},
			expected: 3,
		},
		{
			name: "item exactly at cutoff",
			dataset: []HeatingStationStatus{
				{FetchTime: cutoff.Unix()},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDataset(tt.dataset, cutoff)
			if len(result) != tt.expected {
				t.Errorf("FilterDataset() = %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestComputeIncidentStatistics(t *testing.T) {
	type expectedStats struct {
		avgMonthlyIncidentTimeHours float32
		avgIncidentTimeHours        float32
		maxIncidentTimeHours        float32
		name                        string
	}

	hoursInAMonth := 24.0 * 30.4375

	nowTs := time.Now().Unix()

	tests := []struct {
		name           string
		dataset        []HeatingStationStatus
		expectedValues map[int64]expectedStats
	}{
		{
			name:           "empty dataset",
			dataset:        []HeatingStationStatus{},
			expectedValues: map[int64]expectedStats{},
		},
		{
			name: "single station no incidents",
			dataset: []HeatingStationStatus{
				{GeoId: 1, Name: "Station1", Status: "working", FetchTime: nowTs - 1000},
				{GeoId: 1, Name: "Station1bis", Status: "working", FetchTime: nowTs - 2000},
			},
			expectedValues: map[int64]expectedStats{
				1: {avgMonthlyIncidentTimeHours: 0, avgIncidentTimeHours: 0, maxIncidentTimeHours: 0, name: "Station1"},
			},
		},

		{
			name: "stations with incidents, and should pick last name",
			dataset: []HeatingStationStatus{
				{GeoId: 1, Name: "Station1", Status: "working", FetchTime: nowTs - 6000},
				{GeoId: 1, Name: "Station1boop", Status: "broken", FetchTime: nowTs - 5000},
				{GeoId: 1, Name: "Station1boop", Status: "issues", FetchTime: nowTs - 4500},
				{GeoId: 1, Name: "Station1", Status: "working", FetchTime: nowTs - 4000},
				{GeoId: 1, Name: "Station1", Status: "working", FetchTime: nowTs - 1000},
				{GeoId: 1, Name: "Station1not", Status: "working", FetchTime: nowTs - 1500},
				{GeoId: 2, Name: "Station2", Status: "working", FetchTime: nowTs - 5000},
				{GeoId: 2, Name: "Station2", Status: "broken", FetchTime: nowTs - 2000},
				{GeoId: 2, Name: "Station4", Status: "working", FetchTime: nowTs},
				{GeoId: 3, Name: "Station10", Status: "working", FetchTime: nowTs - 10000},
				{GeoId: 3, Name: "Station10", Status: "broken", FetchTime: nowTs - 5000},
				{GeoId: 10, Name: "Station100", Status: "working", FetchTime: nowTs - 5000},
				{GeoId: 10, Name: "Station100", Status: "issue", FetchTime: nowTs - 4000},
				{GeoId: 10, Name: "Station100", Status: "working", FetchTime: nowTs - 3000},
				{GeoId: 10, Name: "Station100", Status: "issue", FetchTime: nowTs - 2000},
				{GeoId: 10, Name: "Station100", Status: "working", FetchTime: nowTs - 1000},
				{GeoId: 20, Name: "Station200", Status: "working", FetchTime: nowTs - 10000},
				{GeoId: 20, Name: "Station200", Status: "issues", FetchTime: nowTs - 5000},
				{GeoId: 20, Name: "Station200", Status: "working", FetchTime: nowTs - 4000},
				{GeoId: 20, Name: "Station200", Status: "issues", FetchTime: nowTs - 2000},
				{GeoId: 20, Name: "Station200", Status: "working", FetchTime: nowTs - 1500},
				{GeoId: 20, Name: "Station200", Status: "working", FetchTime: nowTs},
			},
			expectedValues: map[int64]expectedStats{
				1: {
					avgMonthlyIncidentTimeHours: float32((1000.0 / 3600.0) / ((5000.0 / 3600.0) / hoursInAMonth)),
					avgIncidentTimeHours:        1000.0 / 3600.0,
					maxIncidentTimeHours:        1000.0 / 3600.0,
					name:                        "Station1",
				},
				2: {
					avgMonthlyIncidentTimeHours: float32((2000.0 / 3600.0) / ((5000.0 / 3600.0) / hoursInAMonth)),
					avgIncidentTimeHours:        2000.0 / 3600.0,
					maxIncidentTimeHours:        2000.0 / 3600.0,
					name:                        "Station4",
				},
				3: {
					avgMonthlyIncidentTimeHours: float32((5000.0 / 3600.0) / ((10000.0 / 3600.0) / hoursInAMonth)),
					avgIncidentTimeHours:        5000.0 / 3600.0,
					maxIncidentTimeHours:        5000.0 / 3600.0,
					name:                        "Station10",
				},
				10: {
					avgMonthlyIncidentTimeHours: float32((2000.0 / 3600.0) / ((4000.0 / 3600.0) / hoursInAMonth)),
					avgIncidentTimeHours:        1000.0 / 3600.0,
					maxIncidentTimeHours:        1000.0 / 3600.0,
					name:                        "Station100",
				},
				20: {
					avgMonthlyIncidentTimeHours: float32((1500.0 / 3600.0) / ((10000.0 / 3600.0) / hoursInAMonth)),
					avgIncidentTimeHours:        750.0 / 3600.0,
					maxIncidentTimeHours:        1000.0 / 3600.0,
					name:                        "Station200",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeIncidentStatistics(tt.dataset)
			if len(result) != len(tt.expectedValues) {
				t.Fatalf("ComputeIncidentStatistics() = %d stations, want %d", len(result), len(tt.expectedValues))
			}

			// Test that output contains exactly the expected stations
			if len(result) != len(tt.expectedValues) {
				t.Fatalf("ComputeIncidentStatistics() returned %d stations, but expected %d stations in expectedValues map", len(result), len(tt.expectedValues))
			}

			// Test expected values for each station
			for _, station := range result {
				if expected, exists := tt.expectedValues[station.GeoId]; exists {
					if abs(float64(station.AvgIncidentTimeHours-expected.avgIncidentTimeHours)) > 0.01 {
						t.Fatalf("Station %d AvgIncidentTimeHours = %.3f, want %.3f", station.GeoId, station.AvgIncidentTimeHours, expected.avgIncidentTimeHours)
					}

					if abs(float64(station.MaxIncidentTimeHours-expected.maxIncidentTimeHours)) > 0.01 {
						t.Fatalf("Station %d MaxIncidentTimeHours = %.3f, want %.3f", station.GeoId, station.MaxIncidentTimeHours, expected.maxIncidentTimeHours)
					}

					if abs(float64(station.AvgMonthlyIncidentTimeHours-expected.avgMonthlyIncidentTimeHours)) > 0.01 {
						t.Fatalf("Station %d AvgMonthlyIncidentTimeHours = %.3f, want %.3f", station.GeoId, station.AvgMonthlyIncidentTimeHours, expected.avgMonthlyIncidentTimeHours)
					}

					if expected.name != "" && station.LastName != expected.name {
						t.Fatalf("Station %d LastName = %s, want %s", station.GeoId, station.LastName, expected.name)
					}

				} else {
					t.Fatalf("Station %d not found in expectedValues map", station.GeoId)
				}
			}

			// Test ranking order (worst first)
			for i := 0; i < len(result)-1; i++ {
				if result[i].AvgMonthlyIncidentTimeHours < result[i+1].AvgMonthlyIncidentTimeHours {
					t.Fatalf("Stations not sorted by AvgMonthlyIncidentTimeHours descending")
				}
				if result[i].Rank != i+1 {
					t.Fatalf("Station rank = %d, want %d", result[i].Rank, i+1)
				}
			}
		})
	}
}
