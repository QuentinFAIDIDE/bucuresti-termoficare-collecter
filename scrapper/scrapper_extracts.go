package scrapper

import (
	"errors"
	"fmt"
	"time"
)

type StationStatesCount struct {
	Time      time.Time `json:"Time"`
	NumGreen  int       `json:"numGreen"`
	NumYellow int       `json:"numYellow"`
	NumRed    int       `json:"numRed"`
}

func (t *TermoficareScrapper) GetStatesCounts() (ssc StationStatesCount, err error) {
	if len(t.rawData) == 0 {
		return ssc, errors.New("no data pulled")
	}
	for _, e := range t.rawData {
		switch e.Category {
		case "verde":
			ssc.NumGreen++
		case "galben":
			ssc.NumYellow++
		case "rosu":
			ssc.NumRed++
		default:
			return ssc, fmt.Errorf("unknown category: %s", e.Category)
		}
	}
	ssc.Time = t.fetchTime
	return ssc, nil
}

func (t *TermoficareScrapper) GetHeatingStations() (states []HeatingStation, err error) {

	if len(t.rawData) == 0 {
		return nil, errors.New("no data pulled")
	}
	states = make([]HeatingStation, 0, len(t.rawData))
	for _, e := range t.rawData {
		states = append(states, e.toHeatingStation())
	}

	return states, nil
}

func (t *TermoficareScrapper) GetHeatingStationsStatuses() (states []HeatingStationStatus, err error) {
	if len(t.rawData) == 0 {
		return nil, errors.New("no data pulled")
	}
	states = make([]HeatingStationStatus, 0, len(t.rawData))
	for _, e := range t.rawData {
		states = append(states, e.toHeatingStationStatus())
	}

	return states, nil
}
