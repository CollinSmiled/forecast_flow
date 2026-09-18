package openmeteo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestOperationalForecastClientFetchOperationalForecast(
	t *testing.T,
) {
	payload := testOperationalPayload()

	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, _ *http.Request) {
			response.Header().Set(
				"Content-Type",
				"application/json",
			)

			if err := json.NewEncoder(response).Encode(payload); err != nil {
				t.Errorf("encode response: %v", err)
			}
		},
	))
	defer server.Close()

	client := newOperationalForecastClient(
		server.URL,
		server.Client(),
	)

	retrievedAt := time.Date(
		2026,
		time.September,
		17,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	client.now = func() time.Time {
		return retrievedAt
	}

	snapshot, err := client.FetchOperationalForecast(
		context.Background(),
		location.Location{
			ID:        3,
			City:      "Jakarta",
			Country:   "Indonesia",
			Latitude:  -6.21462,
			Longitude: 106.84513,
			Timezone:  "Asia/Jakarta",
		},
		10,
	)
	if err != nil {
		t.Fatalf("fetch operational forecast: %v", err)
	}

	if snapshot.LocationID != 3 {
		t.Errorf(
			"location ID = %d, want 3",
			snapshot.LocationID,
		)
	}

	if !snapshot.RetrievedAt.Equal(retrievedAt) {
		t.Errorf(
			"retrieval time = %v, want %v",
			snapshot.RetrievedAt,
			retrievedAt,
		)
	}

	if len(snapshot.Hourly) != 2 {
		t.Errorf(
			"hourly count = %d, want 2",
			len(snapshot.Hourly),
		)
	}

	if len(snapshot.Daily) != 1 {
		t.Errorf(
			"daily count = %d, want 1",
			len(snapshot.Daily),
		)
	}
}

func TestOperationalForecastClientFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/v1/forecast" {
				t.Errorf(
					"path = %q, want %q",
					request.URL.Path,
					"/v1/forecast",
				)
			}

			if request.URL.Query().Get("models") != "best_match" {
				t.Errorf(
					"models = %q, want best_match",
					request.URL.Query().Get("models"),
				)
			}

			response.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = response.Write([]byte(`{
				"latitude": -6.2,
				"longitude": 106.8,
				"generationtime_ms": 0.8,
				"utc_offset_seconds": 25200,
				"timezone": "Asia/Jakarta",
				"timezone_abbreviation": "GMT+7",
				"elevation": 16,
				"current": {
					"time": "2026-09-18T14:15",
					"interval": 900,
					"temperature_2m": 31.2,
					"relative_humidity_2m": 74,
					"apparent_temperature": 35.6,
					"is_day": 1,
					"weather_code": 2
				},
				"hourly": {
					"time": [
						"2026-09-18T14:00",
						"2026-09-18T15:00"
					],
					"temperature_2m": [31.1, 31.5],
					"precipitation_probability": [20, null]
				},
				"daily": {
					"time": ["2026-09-18"],
					"temperature_2m_max": [32.4],
					"temperature_2m_min": [25.1],
					"sunrise": ["2026-09-18T05:44"],
					"sunset": ["2026-09-18T17:51"]
				}
			}`))
		},
	))
	defer server.Close()

	client := newOperationalForecastClient(
		server.URL,
		server.Client(),
	)

	payload, err := client.fetch(
		context.Background(),
		OperationalForecastRequest{
			Latitude:     -6.21462,
			Longitude:    106.84513,
			Timezone:     "Asia/Jakarta",
			ForecastDays: 10,
		},
	)
	if err != nil {
		t.Fatalf("fetch operational forecast: %v", err)
	}

	if payload.Current.Temperature2M == nil ||
		*payload.Current.Temperature2M != 31.2 {
		t.Errorf(
			"current temperature = %v, want 31.2",
			payload.Current.Temperature2M,
		)
	}

	if len(payload.Hourly.PrecipitationProbability) != 2 {
		t.Fatalf(
			"probability count = %d, want 2",
			len(payload.Hourly.PrecipitationProbability),
		)
	}

	if payload.Hourly.PrecipitationProbability[1] != nil {
		t.Errorf(
			"second probability = %v, want nil",
			payload.Hourly.PrecipitationProbability[1],
		)
	}

	if len(payload.Daily.Time) != 1 {
		t.Errorf(
			"daily count = %d, want 1",
			len(payload.Daily.Time),
		)
	}
}

func TestOperationalForecastClientReturnsProviderError(
	t *testing.T,
) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusBadRequest)

			_, _ = response.Write([]byte(`{
				"error": true,
				"reason": "Invalid timezone"
			}`))
		},
	))
	defer server.Close()

	client := newOperationalForecastClient(
		server.URL,
		server.Client(),
	)

	_, err := client.fetch(
		context.Background(),
		OperationalForecastRequest{
			Latitude:     -6.21462,
			Longitude:    106.84513,
			Timezone:     "Asia/Jakarta",
			ForecastDays: 10,
		},
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "Invalid timezone") {
		t.Errorf(
			"error = %q, want provider reason",
			err,
		)
	}
}
