package openmeteo

import (
	"strings"
	"testing"
	"time"
)

func TestBuildVerificationWeatherEndpoint(t *testing.T) {
	endpoint, err := buildVerificationWeatherEndpoint(
		verificationWeatherBaseURL,
		VerificationWeatherRequest{
			Latitude:  -6.21462,
			Longitude: 106.84513,
			Timezone:  "Asia/Jakarta",
			StartDate: time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("build endpoint: %v", err)
	}

	query := endpoint.Query()
	if endpoint.Path != "/v1/archive" {
		t.Errorf("path = %q, want /v1/archive", endpoint.Path)
	}
	if query.Get("start_date") != "2026-09-20" || query.Get("end_date") != "2026-09-21" {
		t.Errorf("date range = %q to %q, want 2026-09-20 to 2026-09-21", query.Get("start_date"), query.Get("end_date"))
	}
	if query.Get("models") != "best_match" {
		t.Errorf("models = %q, want best_match", query.Get("models"))
	}
	if !strings.Contains(query.Get("hourly"), "temperature_2m") {
		t.Errorf("hourly variables = %q, want temperature_2m", query.Get("hourly"))
	}
}

func TestVerificationWeatherRequestRejectsInvalidRange(t *testing.T) {
	request := VerificationWeatherRequest{
		Timezone:  "Asia/Jakarta",
		StartDate: time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC),
	}

	if err := request.validate(); err == nil {
		t.Fatal("invalid range error = nil, want an error")
	}
}

func TestVerificationWeatherRequestRejectsOversizedRange(t *testing.T) {
	startDate := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	request := VerificationWeatherRequest{
		Timezone:  "Asia/Jakarta",
		StartDate: startDate,
		EndDate:   startDate.AddDate(0, 0, maximumVerificationDays),
	}

	if err := request.validate(); err == nil {
		t.Fatal("oversized range error = nil, want an error")
	}
}
