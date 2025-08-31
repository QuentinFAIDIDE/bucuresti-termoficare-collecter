package scrapper

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

type HeatingStation struct {
	GeoId     int64
	Name      string
	Latitude  float64
	Longitude float64
}

type HeatingStationStatus struct {
	GeoId            int64
	Name             string
	FetchTime        time.Time
	Status           string // working,issue,broken
	IncidentType     string // remediare ACC
	IncidentText     string // stare
	EstimatedFixDate time.Time
	Latitude         float64
	Longitude        float64
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
		GeoId:     id,
		Name:      rss.Denumire,
		Latitude:  rss.Latitudine,
		Longitude: rss.Longitudine,
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

func (rss *remoteStreetHeatingStatus) toHeatingStationStatus() HeatingStationStatus {
	id := rss.generateLocationId()
	return HeatingStationStatus{
		GeoId:            id,
		Name:             rss.Denumire,
		FetchTime:        time.Now(), // TO CHANGE
		Status:           rss.getEnglishStatus(),
		IncidentType:     rss.Tip,
		IncidentText:     rss.Stare,
		EstimatedFixDate: time.Now(), // TO CHANGE
		Latitude:         rss.Latitudine,
		Longitude:        rss.Longitudine,
	}
}

type TermoficareScrapper struct {
	httpClient *http.Client
	rawData    []remoteStreetHeatingStatus
	fetchTime  time.Time
}

func NewTermoficareScrapper(proxyUrl string) (*TermoficareScrapper, error) {
	var httpClient *http.Client
	if proxyUrl == "" {
		httpClient = http.DefaultClient
	} else {
		url, err := url.Parse(proxyUrl)
		if err != nil {
			return nil, errors.New("invalid proxy url passed")
		}
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(url),
			},
		}
	}

	return &TermoficareScrapper{
		httpClient: httpClient,
	}, nil
}

func (t *TermoficareScrapper) PullData() (err error) {
	t.fetchTime = time.Now()
	t.rawData, err = t.getStreetHeatingStatuses()
	return err
}

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

func (t *TermoficareScrapper) getStreetHeatingStatuses() ([]remoteStreetHeatingStatus, error) {

	const hartaUrl = "https://www.cmteb.ro/harta_stare_sistem_termoficare_bucuresti.php"

	webpageContent, err := http.Get(hartaUrl)
	if err != nil {
		return nil, err
	}
	defer webpageContent.Body.Close()

	body, err := io.ReadAll(webpageContent.Body)
	if err != nil {
		return nil, err
	}

	return extractStreetStatusesFromPage(string(body), t.fetchTime)
}

func extractStreetStatusesFromPage(webpageContent string, fetchTime time.Time) ([]remoteStreetHeatingStatus, error) {
	var statuses []remoteStreetHeatingStatus

	greenDatumsLine, yellowDatumsLine, redDaturmsLine, err := extractDatumsLines(webpageContent)
	if err != nil {
		return nil, fmt.Errorf("failed to find lines in web page content: %w", err)
	}

	var currentItems []remoteStreetHeatingStatus
	err = json.Unmarshal([]byte(greenDatumsLine), &currentItems)
	if err != nil {
		log.Print(string(greenDatumsLine))
		return nil, fmt.Errorf("failed to parse green streets: %w", err)
	}
	for _, item := range currentItems {
		item.Category = "verde"
		statuses = append(statuses, item)
	}

	err = json.Unmarshal([]byte(yellowDatumsLine), &currentItems)
	if err != nil {
		return nil, fmt.Errorf("failed to parse yellow streets: %w", err)
	}
	for _, item := range currentItems {
		item.Category = "galben"
		statuses = append(statuses, item)
	}

	err = json.Unmarshal([]byte(redDaturmsLine), &currentItems)
	if err != nil {
		return nil, fmt.Errorf("failed to parse red streets: %w", err)
	}
	for _, item := range currentItems {
		item.Category = "rosu"
		statuses = append(statuses, item)
	}

	for i := range statuses {
		statuses[i].Latitudine = math.Round(statuses[i].Latitudine*1e6) / 1e6
		statuses[i].Longitudine = math.Round(statuses[i].Longitudine*1e6) / 1e6
		statuses[i].FetchTime = fetchTime
	}

	return statuses, nil
}

func extractDatumsLines(webpageContent string) (greenDatumsLine, yellowDatumsLine, redDaturmsLine string, err error) {

	const greenDatumsLinePrefix = "var passedFeatures_verde = "
	const yellowDatumsLinePrefix = "var passedFeatures_galben = "
	const redDatumsLinePrefix = "var passedFeatures_rosu = "

	lines := strings.Split(webpageContent, "\n")
	for _, l := range lines {
		l = strings.ReplaceAll(l, "\t", "")
		l = strings.ReplaceAll(l, ";", "")
		if strings.HasPrefix(l, greenDatumsLinePrefix) {
			greenDatumsLine = strings.TrimPrefix(l, greenDatumsLinePrefix)
		} else if strings.HasPrefix(l, yellowDatumsLinePrefix) {
			yellowDatumsLine = strings.TrimPrefix(l, yellowDatumsLinePrefix)
		} else if strings.HasPrefix(l, redDatumsLinePrefix) {
			redDaturmsLine = strings.TrimPrefix(l, redDatumsLinePrefix)
		}
	}

	if greenDatumsLine == "" || yellowDatumsLine == "" || redDaturmsLine == "" {
		return "", "", "", errors.New("could not find datums lines in page")
	}

	return greenDatumsLine, yellowDatumsLine, redDaturmsLine, nil
}
