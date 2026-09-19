package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/hotforecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type stubLatestForecastReader struct {
	locationID int64
	result     hotforecast.Latest
	err        error
}

func (stub *stubLatestForecastReader) GetLatest(
	_ context.Context,
	locationID int64,
) (hotforecast.Latest, error) {
	stub.locationID = locationID

	return stub.result, stub.err
}

func TestGetLatestForecast(t *testing.T) {
	temperature := 31.1
	validAt := time.Date(
		2026,
		time.September,
		19,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	reader := &stubLatestForecastReader{
		result: hotforecast.Latest{
			EventID: "forecast-123",
			Location: location.Location{
				ID:          3,
				City:        "Jakarta",
				Country:     "Indonesia",
				CountryCode: "ID",
			},
			Snapshot: forecast.OperationalForecastSnapshot{
				LocationID:  3,
				Source:      "open_meteo_best_match",
				RetrievedAt: validAt,
				Timezone:    "Asia/Jakarta",
				Current: forecast.CurrentConditions{
					ValidAt:         validAt,
					IntervalSeconds: 900,
					WeatherMetrics: forecast.WeatherMetrics{
						Temperature2M: &temperature,
					},
				},
				Hourly: []forecast.OperationalHourlyForecast{
					{ValidAt: validAt},
				},
				Daily: []forecast.DailyForecast{
					{Date: "2026-09-19"},
				},
			},
		},
	}

	router := http.NewServeMux()
	registerForecastRoutes(router, reader)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/locations/3/forecast",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", response.Code)
	}

	if reader.locationID != 3 {
		t.Errorf("location ID = %d, want 3", reader.locationID)
	}

	var body latestForecastEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Data.Location.City != "Jakarta" {
		t.Errorf(
			"city = %q, want Jakarta",
			body.Data.Location.City,
		)
	}

	if body.Data.Current.Temperature2M == nil ||
		*body.Data.Current.Temperature2M != 31.1 {
		t.Errorf(
			"temperature = %v, want 31.1",
			body.Data.Current.Temperature2M,
		)
	}

	if len(body.Data.Hourly) != 1 || len(body.Data.Daily) != 1 {
		t.Errorf(
			"hourly/daily counts = %d/%d, want 1/1",
			len(body.Data.Hourly),
			len(body.Data.Daily),
		)
	}
}

func TestGetLatestForecastRejectsInvalidLocationID(t *testing.T) {
	router := http.NewServeMux()
	registerForecastRoutes(router, &stubLatestForecastReader{})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/locations/not-a-number/forecast",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want 400",
			response.Code,
		)
	}
}

func TestGetLatestForecastReturnsNotFound(t *testing.T) {
	reader := &stubLatestForecastReader{
		err: fmt.Errorf(
			"read forecast: %w",
			hotforecast.ErrLatestForecastNotFound,
		),
	}
	router := http.NewServeMux()
	registerForecastRoutes(router, reader)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/locations/99/forecast",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"status code = %d, want 404",
			response.Code,
		)
	}
}
