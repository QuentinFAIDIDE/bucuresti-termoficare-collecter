package scrapper

import (
	"testing"
	"time"
)

func TestGetStatesCounts(t *testing.T) {
	tests := []struct {
		name        string
		rawData     []remoteStreetHeatingStatus
		wantGreen   int
		wantYellow  int
		wantRed     int
		wantErr     bool
		errContains string
	}{
		{
			name:        "no data",
			rawData:     nil,
			wantErr:     true,
			errContains: "no data pulled",
		},
		{
			name:        "empty data",
			rawData:     []remoteStreetHeatingStatus{},
			wantErr:     true,
			errContains: "no data pulled",
		},
		{
			name: "all green",
			rawData: []remoteStreetHeatingStatus{
				{Category: "verde"},
				{Category: "verde"},
			},
			wantGreen: 2,
		},
		{
			name: "all yellow",
			rawData: []remoteStreetHeatingStatus{
				{Category: "galben"},
				{Category: "galben"},
				{Category: "galben"},
			},
			wantYellow: 3,
		},
		{
			name: "all red",
			rawData: []remoteStreetHeatingStatus{
				{Category: "rosu"},
			},
			wantRed: 1,
		},
		{
			name: "mixed categories",
			rawData: []remoteStreetHeatingStatus{
				{Category: "verde"},
				{Category: "galben"},
				{Category: "rosu"},
				{Category: "verde"},
				{Category: "galben"},
			},
			wantGreen:  2,
			wantYellow: 2,
			wantRed:    1,
		},
		{
			name: "unknown category",
			rawData: []remoteStreetHeatingStatus{
				{Category: "verde"},
				{Category: "unknown"},
			},
			wantErr:     true,
			errContains: "unknown category: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scrapper := &TermoficareScrapper{
				rawData: tt.rawData,
			}

			ssc, err := scrapper.GetStatesCounts()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Fatalf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
				return
			}

			if ssc.NumGreen != tt.wantGreen {
				t.Fatalf("green count = %d, want %d", ssc.NumGreen, tt.wantGreen)
			}
			if ssc.NumYellow != tt.wantYellow {
				t.Fatalf("yellow count = %d, want %d", ssc.NumYellow, tt.wantYellow)
			}
			if ssc.NumRed != tt.wantRed {
				t.Fatalf("red count = %d, want %d", ssc.NumRed, tt.wantRed)
			}
		})
	}
}

func TestGetHeatingStations(t *testing.T) {
	tests := []struct {
		name    string
		rawData []remoteStreetHeatingStatus
		wantLen int
		wantErr bool
	}{
		{
			name:    "no data",
			rawData: nil,
			wantErr: true,
		},
		{
			name:    "empty data",
			rawData: []remoteStreetHeatingStatus{},
			wantErr: true,
		},
		{
			name: "single station",
			rawData: []remoteStreetHeatingStatus{
				{
					Stare:       "StareTest",
					Denumire:    "Test Station",
					Tip:         "Tip test",
					Remediere:   "10.09.2025 10:30",
					Latitudine:  44.4267,
					Longitudine: 26.1025,
					Category:    "verde",
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple stations",
			rawData: []remoteStreetHeatingStatus{
				{
					Denumire:    "Station 1",
					Latitudine:  44.4267,
					Longitudine: 26.1025,
					Category:    "verde",
				},
				{
					Denumire:    "Station 2",
					Latitudine:  44.4300,
					Longitudine: 26.1100,
					Category:    "verde",
				},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scrapper := &TermoficareScrapper{
				rawData:   tt.rawData,
				fetchTime: time.Now(),
			}

			stations, err := scrapper.GetHeatingStations()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				if stations != nil {
					t.Errorf("expected nil stations but got %v", stations)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(stations) != tt.wantLen {
				t.Errorf("stations length = %d, want %d", len(stations), tt.wantLen)
			}

			for i, station := range stations {
				if station.Name != tt.rawData[i].Denumire {
					t.Errorf("station[%d].Name = %q, want %q", i, station.Name, tt.rawData[i].Denumire)
				}
				if station.Latitude != tt.rawData[i].Latitudine {
					t.Errorf("station[%d].Latitude = %f, want %f", i, station.Latitude, tt.rawData[i].Latitudine)
				}
				if station.Longitude != tt.rawData[i].Longitudine {
					t.Errorf("station[%d].Longitude = %f, want %f", i, station.Longitude, tt.rawData[i].Longitudine)
				}
			}
		})
	}
}

func TestToHeatingStationStatus(t *testing.T) {
	tests := []struct {
		name           string
		input          remoteStreetHeatingStatus
		wantName       string
		wantStatus     string
		wantHasFixDate bool
	}{
		{
			name: "verde category with remediere date",
			input: remoteStreetHeatingStatus{
				Denumire:    "Test Station",
				Category:    "verde",
				Tip:         "Maintenance",
				Stare:       "Working normally",
				Remediere:   "05.09.2025 20:00",
				Latitudine:  44.4267,
				Longitudine: 26.1025,
				FetchTime:   time.Now(),
			},
			wantName:       "Test Station",
			wantStatus:     "working",
			wantHasFixDate: true,
		},
		{
			name: "galben category without remediere date",
			input: remoteStreetHeatingStatus{
				Denumire:    "Yellow Station",
				Category:    "galben",
				Tip:         "Issue",
				Stare:       "Minor problem",
				Remediere:   "",
				Latitudine:  44.4300,
				Longitudine: 26.1100,
				FetchTime:   time.Now(),
			},
			wantName:       "Yellow Station",
			wantStatus:     "issue",
			wantHasFixDate: false,
		},
		{
			name: "rosu category with remediere date",
			input: remoteStreetHeatingStatus{
				Denumire:    "Red Station",
				Category:    "rosu",
				Tip:         "Broken",
				Stare:       "Not working",
				Remediere:   "01.09.2025 12:00",
				Latitudine:  44.4400,
				Longitudine: 26.1200,
				FetchTime:   time.Now(),
			},
			wantName:       "Red Station",
			wantStatus:     "broken",
			wantHasFixDate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.toHeatingStationStatus()

			if result.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
			}
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", result.Status, tt.wantStatus)
			}
			if result.IncidentType != tt.input.Tip {
				t.Errorf("IncidentType = %q, want %q", result.IncidentType, tt.input.Tip)
			}
			if result.IncidentText != tt.input.Stare {
				t.Errorf("IncidentText = %q, want %q", result.IncidentText, tt.input.Stare)
			}
			if result.Latitude != tt.input.Latitudine {
				t.Errorf("Latitude = %f, want %f", result.Latitude, tt.input.Latitudine)
			}
			if result.Longitude != tt.input.Longitudine {
				t.Errorf("Longitude = %f, want %f", result.Longitude, tt.input.Longitudine)
			}
			if tt.wantHasFixDate && result.EstimatedFixDate == 0 {
				t.Errorf("Expected EstimatedFixDate to be set but got zero time")
			}
			if !tt.wantHasFixDate && result.EstimatedFixDate != 0 {
				t.Errorf("Expected EstimatedFixDate to be zero but got %v", result.EstimatedFixDate)
			}
		})
	}
}
