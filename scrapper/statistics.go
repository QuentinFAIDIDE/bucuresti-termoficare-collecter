package scrapper

import (
	"slices"
	"time"
)

type StationIncidentStatsDbRow struct {
	City                        string  `json:"city" dynamodbav:"City"`
	Rank                        int     `json:"rank" dynamodbav:"Rank"`
	GeoId                       int64   `json:"geoId" dynamodbav:"GeoId"`
	LastName                    string  `json:"lastName" dynamodbav:"LastName"`
	Latitude                    float64 `json:"latitude" dynamodbav:"Latitude"`
	Longitude                   float64 `json:"longitude" dynamodbav:"Longitude"`
	AvgMonthlyIncidentTimeHours float32 `json:"avgMonthlyIncidentTimeHours" dynamodbav:"AvgMonthlyIncidentTimeHours"`
	AvgIncidentTimeHours        float32 `json:"avgIncidentTimeHours" dynamodbav:"AvgIncidentTimeHours"`
	MaxIncidentTimeHours        float32 `json:"maxIncidentTimeHours" dynamodbav:"MaxIncidentTimeHours"`
}

type StationIncidentsData struct {
	GeoId                   int64
	IncidentsDurationsHours []float64
	Name                    string
	Latitude                float64
	Longitude               float64
	FirstDate               int64
	LastDate                int64
}

func FilterDataset(dataset []HeatingStationStatus, cutoffTimestamp time.Time) []HeatingStationStatus {
	filteredDataset := make([]HeatingStationStatus, 0, len(dataset))

	for _, item := range dataset {
		fetchTime := time.Unix(item.FetchTime, 0)
		if fetchTime.After(cutoffTimestamp) {
			filteredDataset = append(filteredDataset, item)
		}
	}

	return filteredDataset
}

func ComputeIncidentStatistics(dataset []HeatingStationStatus) []StationIncidentStatsDbRow {
	stations := make([]StationIncidentStatsDbRow, 0, 1024)

	slices.SortFunc(dataset, func(a, b HeatingStationStatus) int {
		return int(a.FetchTime - b.FetchTime)
	})

	stationsIncidentStats := computeIncidentsPerStation(dataset)

	for _, stats := range stationsIncidentStats {
		stations = append(stations, aggregateIncidentDurations(stats))
	}

	slices.SortFunc(stations, func(a, b StationIncidentStatsDbRow) int {
		if a.AvgMonthlyIncidentTimeHours > b.AvgMonthlyIncidentTimeHours {
			return -1
		}
		if a.AvgMonthlyIncidentTimeHours < b.AvgMonthlyIncidentTimeHours {
			return 1
		}
		return 0
	})

	for i := range stations {
		stations[i].Rank = i + 1
	}

	return stations
}

func aggregateIncidentDurations(stats StationIncidentsData) StationIncidentStatsDbRow {
	const (
		avgDaysPerMonth = 30.4375
		hoursPerDay     = 24.0
		secondsPerHour  = 3600.0
	)

	var (
		totalIncidentsHours  float64
		maxIncidentTimeHours float64
		rangeDurationHours   float64
		rangeDurationMonths  float64
		numIncidents         float64
	)

	rangeDurationHours = (float64(stats.LastDate) - float64(stats.FirstDate)) / secondsPerHour
	rangeDurationMonths = rangeDurationHours / (hoursPerDay * avgDaysPerMonth)
	numIncidents = float64(len(stats.IncidentsDurationsHours))

	for i := range stats.IncidentsDurationsHours {
		totalIncidentsHours += stats.IncidentsDurationsHours[i]
		if stats.IncidentsDurationsHours[i] > maxIncidentTimeHours {
			maxIncidentTimeHours = stats.IncidentsDurationsHours[i]
		}
	}

	return StationIncidentStatsDbRow{
		City:                        "Bucharest",
		GeoId:                       stats.GeoId,
		LastName:                    stats.Name,
		Latitude:                    stats.Latitude,
		Longitude:                   stats.Longitude,
		AvgMonthlyIncidentTimeHours: float32(totalIncidentsHours) / float32(rangeDurationMonths),
		MaxIncidentTimeHours:        float32(maxIncidentTimeHours),
		AvgIncidentTimeHours:        float32(totalIncidentsHours / numIncidents),
	}
}

func computeIncidentsPerStation(dataset []HeatingStationStatus) map[int64]StationIncidentsData {
	lastStationIncident := make(map[int64]int64, 1024)
	stationsIncidentData := make(map[int64]StationIncidentsData, 1024)

	for _, row := range dataset {
		if _, exists := stationsIncidentData[row.GeoId]; !exists {
			stationsIncidentData[row.GeoId] = StationIncidentsData{
				GeoId:                   row.GeoId,
				Name:                    row.Name,
				Latitude:                row.Latitude,
				Longitude:               row.Longitude,
				IncidentsDurationsHours: make([]float64, 0, 128),
			}
		}

		if row.Name != stationsIncidentData[row.GeoId].Name {
			stats := stationsIncidentData[row.GeoId]
			stats.Name = row.Name
			stationsIncidentData[row.GeoId] = stats
		}

		if stationsIncidentData[row.GeoId].FirstDate == 0 || stationsIncidentData[row.GeoId].FirstDate > row.FetchTime {
			stats := stationsIncidentData[row.GeoId]
			stats.FirstDate = row.FetchTime
			stationsIncidentData[row.GeoId] = stats
		}

		if stationsIncidentData[row.GeoId].LastDate == 0 || stationsIncidentData[row.GeoId].LastDate < row.FetchTime {
			stats := stationsIncidentData[row.GeoId]
			stats.LastDate = row.FetchTime
			stationsIncidentData[row.GeoId] = stats
		}

		lastIncidentTime, hadIncident := lastStationIncident[row.GeoId]
		stationIsInIncident := row.Status != "working"
		stationWasInIncident := (hadIncident && lastIncidentTime != 0)

		if stationIsInIncident && !stationWasInIncident {
			lastStationIncident[row.GeoId] = row.FetchTime
		}

		if !stationIsInIncident && stationWasInIncident {
			stationStats := stationsIncidentData[row.GeoId]
			stationStats.IncidentsDurationsHours = append(stationStats.IncidentsDurationsHours, float64(row.FetchTime-lastIncidentTime)/3600.0)
			stationsIncidentData[row.GeoId] = stationStats
			lastStationIncident[row.GeoId] = 0
		}
	}

	nowUnix := time.Now().Unix()
	for geoId, lastIncidentTime := range lastStationIncident {
		if lastIncidentTime != 0 {
			stationStats := stationsIncidentData[geoId]
			stationStats.IncidentsDurationsHours = append(stationStats.IncidentsDurationsHours, float64(nowUnix-lastIncidentTime)/3600.0)
			stationStats.LastDate = nowUnix
			stationsIncidentData[geoId] = stationStats
		}
	}

	return stationsIncidentData
}
