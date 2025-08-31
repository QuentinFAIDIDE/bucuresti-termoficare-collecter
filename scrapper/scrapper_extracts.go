package scrapper

import (
	"errors"
	"fmt"
)

func (t *TermoficareScrapper) GetStatesCounts() (numGreen int, numYellow int, numRed int, err error) {
	if len(t.rawData) == 0 {
		return 0, 0, 0, errors.New("no data pulled")
	}
	for _, e := range t.rawData {
		switch e.Category {
		case "verde":
			numGreen++
		case "galben":
			numYellow++
		case "rosu":
			numRed++
		default:
			return 0, 0, 0, fmt.Errorf("unknown category: %s", e.Category)
		}
	}
	return numGreen, numYellow, numRed, nil
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
