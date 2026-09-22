package openmeteo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestVerificationWeatherClientFetchVerificationWeather(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/v1/archive" {
				t.Errorf("path = %q, want /v1/archive", request.URL.Path)
			}
			response.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(response).Encode(testVerificationPayload()); err != nil {
				t.Errorf("encode response: %v", err)
			}
		},
	))
	defer server.Close()

	client := newVerificationWeatherClient(server.URL, server.Client())
	retrievedAt := time.Date(2026, time.September, 22, 10, 0, 0, 0, time.UTC)
	client.now = func() time.Time { return retrievedAt }

	snapshot, err := client.FetchVerificationWeather(
		context.Background(),
		location.Location{
			ID:        3,
			Latitude:  -6.21462,
			Longitude: 106.84513,
			Timezone:  "Asia/Jakarta",
		},
		time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch verification weather: %v", err)
	}

	if snapshot.LocationID != 3 {
		t.Errorf("location ID = %d, want 3", snapshot.LocationID)
	}
	if !snapshot.RetrievedAt.Equal(retrievedAt) {
		t.Errorf("retrieved at = %v, want %v", snapshot.RetrievedAt, retrievedAt)
	}
	if len(snapshot.Hourly) != 2 {
		t.Errorf("hourly count = %d, want 2", len(snapshot.Hourly))
	}
}

func TestVerificationWeatherClientRejectsMissingHourlyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, _ *http.Request) {
			response.Header().Set("Content-Type", "application/json")
			_, _ = response.Write([]byte(`{"timezone":"Asia/Jakarta","hourly":{}}`))
		},
	))
	defer server.Close()

	client := newVerificationWeatherClient(server.URL, server.Client())
	date := time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC)
	_, err := client.fetch(context.Background(), VerificationWeatherRequest{
		Timezone:  "Asia/Jakarta",
		StartDate: date,
		EndDate:   date,
	})
	if err == nil {
		t.Fatal("missing hourly data error = nil, want an error")
	}
}
