package scrapper

import (
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
	t.fetchTime = time.Now().UTC()
	t.rawData, err = t.getStreetHeatingStatuses()
	return err
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
