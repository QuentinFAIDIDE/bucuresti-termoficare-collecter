package scrapper

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestExtractStreetStatusesFromPage(t *testing.T) {

	content, err := os.ReadFile("test_data")
	if err != nil {
		t.Fatal(err)
	}

	got, err := extractStreetStatusesFromPage(string(content))
	if err != nil {
		t.Fatalf("extractStreetStatusesFromPage() error = %v", err)
	}

	if len(got) < 2 {
		t.Fatalf("extractStreetStatusesFromPage() returned too few results, got %v entries", len(got))
	}

	var firstGreenExpected remoteStreetHeatingStatus
	err = json.Unmarshal([]byte(`
		{"stare":"Functionare normala","culoare":"#008217","denumire":"1 C3\/1","longitudine":26.164120976753,"latitudine":44.432576444788,"tip":"-"}`,
	), &firstGreenExpected)
	if err != nil {
		t.Fatalf("parse error = %v", err)
	}
	firstGreenExpected.Category = "verde"
	if !reflect.DeepEqual(firstGreenExpected, got[0]) {
		t.Fatalf("extractStreetStatusesFromPage() returned unexpected first entry, got %+v, expected %+v", got[0], firstGreenExpected)
	}

	i := 1
	for i < len(got) && got[i].Category == "verde" {
		i++
	}
	if i >= len(got) {
		t.Fatalf("extractStreetStatusesFromPage() returned no entry with category 'galben'")
	}

	var firstYellowExpected remoteStreetHeatingStatus
	err = json.Unmarshal([]byte(`
		{"stare":"Functionare deficitara a apei calde de consum din cauza manevrelor de echilibrare hidraulica a retelei termice","culoare":"#ffe53e","denumire":"1 Colentina","longitudine":26.12367664304,"latitudine":44.452127590305,"tip":"Deficienta ACC","remediere":"01.09.2025 12:00"}`,
	), &firstYellowExpected)
	if err != nil {
		t.Fatalf("parse error = %v", err)
	}
	firstYellowExpected.Category = "galben"
	if !reflect.DeepEqual(firstYellowExpected, got[i]) {
		t.Fatalf("extractStreetStatusesFromPage() returned unexpected first yellow entry, got %+v, expected %+v", got[i], firstYellowExpected)
	}

	for i < len(got) && got[i].Category == "galben" {
		i++
	}
	if i >= len(got) {
		t.Fatalf("extractStreetStatusesFromPage() returned no entry with category 'rosu'")
	}

	var firstRedExpected remoteStreetHeatingStatus
	err = json.Unmarshal([]byte(`
		{"stare":"Avarie apa calda  \/ CD - Apartine asociatiei","culoare":"#e0002b","denumire":"1 Stoian Militaru","longitudine":26.098421911171,"latitudine":44.399987462908,"tip":"Oprire ACC","remediere":"05.09.2025 20:00"}`,
	), &firstRedExpected)
	if err != nil {
		t.Fatalf("parse error = %v", err)
	}
	firstRedExpected.Category = "rosu"
	if !reflect.DeepEqual(firstRedExpected, got[i]) {
		t.Fatalf("extractStreetStatusesFromPage() returned unexpected first red entry, got %+v, expected %+v", got[i], firstRedExpected)
	}

	for i < len(got) && got[i].Category == "rosu" {
		i++
	}
	if i < len(got) {
		t.Fatalf("extractStreetStatusesFromPage() returned too many entries with category 'rosu'")
	}
}

func TestSmokeScrapWebsite(t *testing.T) {
	c, err := NewTermoficareScrapper("")
	if err != nil {
		t.Fatalf("NewTermoficareScrapper() error = %v", err)
	}
	_, err = c.GetStreetHeatingStatuses()
	if err != nil {
		t.Fatal(err)
	}
}
