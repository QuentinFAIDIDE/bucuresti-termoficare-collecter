package scrapper

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

var bucharestTz *time.Location
var tzMutex sync.RWMutex

// RemoteStreetHeatingSremoteStreetHeatingStatustatus is a structure that mirrors
// the object listed in the termoficare harta website source.
type remoteStreetHeatingStatus struct {
	Stare       string  `json:"stare"`
	Culoare     string  `json:"culoare"`
	Denumire    string  `json:"denumire"`
	Tip         string  `json:"tip"`
	Remediere   string  `json:"remediere"`
	Longitudine float64 `json:"longitudine"`
	Latitudine  float64 `json:"latitudine"`
	Category    string  // verde, galben, rosu
	FetchTime   time.Time
}

// requests per table:
// get all heating stations: table of heating stations partitioned by city, sortkey geoId, with denumire, longitude and latitude
// get history for one station: table of history for one station, partition key geoId, sort key timestamp descending, storing state, category, remediere, tip

type StationStatesCount struct {
	Time      int64 `json:"time" dynamodbav:"Timestamp"`
	NumGreen  int   `json:"numGreen" dynamodbav:"numGreen"`
	NumYellow int   `json:"numYellow" dynamodbav:"numYellow"`
	NumRed    int   `json:"numRed" dynamodbav:"numRed"`
}

type HeatingStation struct {
	GeoId      int64   `json:"geoId" dynamodbav:"GeoId"`
	Name       string  `json:"name" dynamodbav:"Name"`
	Latitude   float64 `json:"latitude" dynamodbav:"Latitude"`
	Longitude  float64 `json:"longitude" dynamodbav:"Longitude"`
	LastStatus string  `json:"lastStatus" dynamodbav:"LastStatus"` // working,issue,broken
}

type HeatingStationStatus struct {
	GeoId            int64   `json:"geoId" dynamodbav:"GeoId"`
	Name             string  `json:"name" dynamodbav:"Name"`
	FetchTime        int64   `json:"fetchTime" dynamodbav:"Timestamp"`
	Status           string  `json:"status" dynamodbav:"Status"`             // working,issue,broken
	IncidentType     string  `json:"incidentType" dynamodbav:"IncidentType"` // remediare ACC
	IncidentText     string  `json:"incidentText" dynamodbav:"IncidentText"` // stare
	EstimatedFixDate int64   `json:"estimatedFixDate" dynamodbav:"EstimatedFixDate"`
	Latitude         float64 `json:"latitude" dynamodbav:"Latitude"`
	Longitude        float64 `json:"longitude" dynamodbav:"Longitude"`
}

func (rss *remoteStreetHeatingStatus) generateLocationId() int64 {
	locationStr := fmt.Sprintf("%.6f,%.6f", rss.Latitudine, rss.Longitudine)

	h := md5.New()
	io.WriteString(h, locationStr)
	hash := int64(binary.BigEndian.Uint64(h.Sum(nil)))

	return hash
}

func (rss *remoteStreetHeatingStatus) toHeatingStation() HeatingStation {
	id := rss.generateLocationId()
	return HeatingStation{
		GeoId:      id,
		Name:       rss.Denumire,
		Latitude:   rss.Latitudine,
		Longitude:  rss.Longitudine,
		LastStatus: rss.getEnglishStatus(),
	}
}

func (rss *remoteStreetHeatingStatus) getEnglishStatus() string {
	switch rss.Category {
	case "verde":
		return "working"
	case "galben":
		return "issue"
	case "rosu":
		return "broken"
	default:
		panic("unknown category: " + rss.Category)
	}
}

func ensureLocationIsSet() {
	tzMutex.RLock()
	isSet := bucharestTz != nil
	tzMutex.RUnlock()
	if isSet {
		return
	}

	tzMutex.Lock()
	defer tzMutex.Unlock()

	if bucharestTz == nil {
		bucharestTz, _ = time.LoadLocation("Europe/Bucharest")
	}
}

func (rss *remoteStreetHeatingStatus) toHeatingStationStatus() HeatingStationStatus {

	const dateLayout = "02.01.2006 15:04"

	id := rss.generateLocationId()

	ensureLocationIsSet()

	var t time.Time
	var fixDateUnix int64
	if rss.Remediere != "" {
		var err error
		tzMutex.RLock()
		t, err = time.ParseInLocation(dateLayout, rss.Remediere, bucharestTz)
		tzMutex.RUnlock()
		if err != nil {
			log.Fatalf("failed to parse date: %v", err)
		}
		fixDateUnix = t.Unix()
	}

	return HeatingStationStatus{
		GeoId:            id,
		Name:             rss.Denumire,
		FetchTime:        rss.FetchTime.Unix(),
		Status:           rss.getEnglishStatus(),
		IncidentType:     rss.Tip,
		IncidentText:     rss.Stare,
		EstimatedFixDate: fixDateUnix,
		Latitude:         rss.Latitudine,
		Longitude:        rss.Longitudine,
	}
}
