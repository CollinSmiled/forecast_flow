package openmeteo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestSingleRunClientFetchForecastRun(t *testing.T) {
	payload := testHourlyPayload()
	payload.Daily = testDailyPayload().Daily

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

	client := newSingleRunClient(
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

	selectedLocation := location.Location{
		ID:        3,
		City:      "Jakarta",
		Country:   "Indonesia",
		Latitude:  -6.21462,
		Longitude: 106.84513,
		Timezone:  "Asia/Jakarta",
	}

	model := forecast.Model{
		ID:           "ecmwf_ifs",
		Name:         "ECMWF IFS",
		Provider:     "ECMWF",
		ResolutionKM: 9,
	}

	runAt := time.Date(
		2026,
		time.September,
		17,
		6,
		0,
		0,
		0,
		time.UTC,
	)

	result, err := client.FetchForecastRun(
		context.Background(),
		selectedLocation,
		model,
		runAt,
		10,
	)
	if err != nil {
		t.Fatalf("fetch forecast run: %v", err)
	}

	if result.Run.LocationID != 3 {
		t.Errorf(
			"location ID = %d, want 3",
			result.Run.LocationID,
		)
	}

	if result.Run.Model.ID != "ecmwf_ifs" {
		t.Errorf(
			"model ID = %q, want ecmwf_ifs",
			result.Run.Model.ID,
		)
	}

	if !result.Run.ForecastRunAt.Equal(runAt) {
		t.Errorf(
			"forecast run time = %v, want %v",
			result.Run.ForecastRunAt,
			runAt,
		)
	}

	if !result.Run.RetrievedAt.Equal(retrievedAt) {
		t.Errorf(
			"retrieval time = %v, want %v",
			result.Run.RetrievedAt,
			retrievedAt,
		)
	}

	if len(result.Hourly) != 2 {
		t.Errorf(
			"hourly count = %d, want 2",
			len(result.Hourly),
		)
	}

	if len(result.Daily) != 1 {
		t.Errorf(
			"daily count = %d, want 1",
			len(result.Daily),
		)
	}
}

func TestSingleRunClientFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/v1/forecast" {
				t.Errorf(
					"path = %q, want %q",
					request.URL.Path,
					"/v1/forecast",
				)
			}

			if request.Header.Get("Accept") != "application/json" {
				t.Errorf(
					"Accept = %q, want application/json",
					request.Header.Get("Accept"),
				)
			}

			query := request.URL.Query()

			if query.Get("models") != "ecmwf_ifs" {
				t.Errorf(
					"models = %q, want ecmwf_ifs",
					query.Get("models"),
				)
			}

			if query.Get("run") != "2026-09-17T06:00" {
				t.Errorf(
					"run = %q, want 2026-09-17T06:00",
					query.Get("run"),
				)
			}

			response.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = response.Write([]byte(`{
				"latitude": -6.2,
				"longitude": 106.8,
				"generationtime_ms": 1.25,
				"utc_offset_seconds": 25200,
				"timezone": "Asia/Jakarta",
				"timezone_abbreviation": "GMT+7",
				"elevation": 16,
				"hourly": {
					"time": [
						"2026-09-17T13:00",
						"2026-09-17T14:00"
					],
					"temperature_2m": [30.4, null],
					"precipitation_probability": [75, 60],
					"is_day": [1, 1]
				},
				"daily": {
					"time": ["2026-09-17"],
					"temperature_2m_max": [32.1],
					"temperature_2m_min": [25.4],
					"sunrise": ["2026-09-17T05:45"],
					"sunset": ["2026-09-17T17:51"]
				}
			}`))
		},
	))
	defer server.Close()

	client := newSingleRunClient(
		server.URL,
		server.Client(),
	)

	payload, err := client.fetch(
		context.Background(),
		testSingleRunRequest(),
	)
	if err != nil {
		t.Fatalf("fetch single run: %v", err)
	}

	if payload.Timezone != "Asia/Jakarta" {
		t.Errorf(
			"timezone = %q, want Asia/Jakarta",
			payload.Timezone,
		)
	}

	if len(payload.Hourly.Time) != 2 {
		t.Fatalf(
			"hourly count = %d, want 2",
			len(payload.Hourly.Time),
		)
	}

	if payload.Hourly.Temperature2M[0] == nil ||
		*payload.Hourly.Temperature2M[0] != 30.4 {
		t.Errorf(
			"first temperature = %v, want 30.4",
			payload.Hourly.Temperature2M[0],
		)
	}

	if payload.Hourly.Temperature2M[1] != nil {
		t.Errorf(
			"second temperature = %v, want nil",
			payload.Hourly.Temperature2M[1],
		)
	}

	if len(payload.Daily.Time) != 1 {
		t.Fatalf(
			"daily count = %d, want 1",
			len(payload.Daily.Time),
		)
	}
}

func TestSingleRunClientReturnsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusBadRequest)

			_, _ = response.Write([]byte(`{
				"error": true,
				"reason": "No data is available for this run."
			}`))
		},
	))
	defer server.Close()

	client := newSingleRunClient(
		server.URL,
		server.Client(),
	)

	_, err := client.fetch(
		context.Background(),
		testSingleRunRequest(),
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(
		err.Error(),
		"No data is available for this run.",
	) {
		t.Errorf(
			"error = %q, want provider reason",
			err,
		)
	}
}

func testSingleRunRequest() SingleRunRequest {
	return SingleRunRequest{
		Latitude:  -6.21462,
		Longitude: 106.84513,
		Timezone:  "Asia/Jakarta",
		ModelID:   "ecmwf_ifs",
		ForecastRunAt: time.Date(
			2026,
			time.September,
			17,
			6,
			0,
			0,
			0,
			time.UTC,
		),
		ForecastDays: 10,
	}
}
