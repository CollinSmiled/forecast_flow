package hotforecast_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/hotforecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/CollinSmiled/forecast_flow/internal/platform/postgres"
)

func TestRepositoryReplacesOnlyWithNewerForecast(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}

	ctx := context.Background()
	pool, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()

	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin test transaction: %v", err)
	}
	defer func() {
		if err := transaction.Rollback(ctx); err != nil {
			t.Errorf("rollback test transaction: %v", err)
		}
	}()

	locationRepository := location.NewRepository(transaction)
	selectedLocation, err := locationRepository.Upsert(
		ctx,
		location.Location{
			OpenMeteoLocationID: 9_999_999_002,
			City:                "Hot Forecast Test City",
			Country:             "Indonesia",
			CountryCode:         "ID",
			Latitude:            -6.2,
			Longitude:           106.8,
			Timezone:            "Asia/Jakarta",
		},
	)
	if err != nil {
		t.Fatalf("create test location: %v", err)
	}

	repository := hotforecast.NewRepository(transaction)
	retrievedAt := time.Date(
		2026,
		time.September,
		19,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	first := latestForecastEvent(
		t,
		selectedLocation.ID,
		retrievedAt,
		30,
	)

	replaced, err := repository.ReplaceLatest(ctx, first)
	if err != nil {
		t.Fatalf("store first forecast: %v", err)
	}
	if !replaced {
		t.Fatal("first forecast was not stored")
	}

	replaced, err = repository.ReplaceLatest(ctx, first)
	if err != nil {
		t.Fatalf("store duplicate forecast: %v", err)
	}
	if replaced {
		t.Fatal("duplicate forecast replaced current state")
	}

	older := latestForecastEvent(
		t,
		selectedLocation.ID,
		retrievedAt.Add(-time.Hour),
		29,
	)
	replaced, err = repository.ReplaceLatest(ctx, older)
	if err != nil {
		t.Fatalf("store older forecast: %v", err)
	}
	if replaced {
		t.Fatal("older forecast replaced current state")
	}

	newer := latestForecastEvent(
		t,
		selectedLocation.ID,
		retrievedAt.Add(time.Hour),
		31.5,
	)
	replaced, err = repository.ReplaceLatest(ctx, newer)
	if err != nil {
		t.Fatalf("store newer forecast: %v", err)
	}
	if !replaced {
		t.Fatal("newer forecast did not replace current state")
	}

	var (
		storedEventID string
		temperature   float64
		hourlyCount   int
		dailyCount    int
	)

	err = transaction.QueryRow(
		ctx,
		`SELECT event_id, current_temperature_2m
		 FROM public.latest_operational_forecasts
		 WHERE location_id = $1`,
		selectedLocation.ID,
	).Scan(&storedEventID, &temperature)
	if err != nil {
		t.Fatalf("read stored current forecast: %v", err)
	}

	if storedEventID != newer.EventID {
		t.Errorf(
			"stored event ID = %q, want %q",
			storedEventID,
			newer.EventID,
		)
	}

	if temperature != 31.5 {
		t.Errorf("temperature = %v, want 31.5", temperature)
	}

	if err := transaction.QueryRow(
		ctx,
		"SELECT count(*) FROM public.latest_operational_hourly WHERE location_id = $1",
		selectedLocation.ID,
	).Scan(&hourlyCount); err != nil {
		t.Fatalf("count hourly rows: %v", err)
	}

	if err := transaction.QueryRow(
		ctx,
		"SELECT count(*) FROM public.latest_operational_daily WHERE location_id = $1",
		selectedLocation.ID,
	).Scan(&dailyCount); err != nil {
		t.Fatalf("count daily rows: %v", err)
	}

	if hourlyCount != 1 {
		t.Errorf("hourly row count = %d, want 1", hourlyCount)
	}

	if dailyCount != 1 {
		t.Errorf("daily row count = %d, want 1", dailyCount)
	}
}

func latestForecastEvent(
	t *testing.T,
	locationID int64,
	retrievedAt time.Time,
	temperature float64,
) event.LatestForecastEventV1 {
	t.Helper()

	snapshot := forecast.OperationalForecastSnapshot{
		LocationID:  locationID,
		Source:      "open_meteo_best_match",
		RetrievedAt: retrievedAt,
		Timezone:    "Asia/Jakarta",
		Current: forecast.CurrentConditions{
			ValidAt:         retrievedAt,
			IntervalSeconds: 900,
			WeatherMetrics: forecast.WeatherMetrics{
				Temperature2M: &temperature,
			},
		},
		Hourly: []forecast.OperationalHourlyForecast{
			{
				ValidAt: retrievedAt.Add(time.Hour),
				WeatherMetrics: forecast.WeatherMetrics{
					Temperature2M: &temperature,
				},
			},
		},
		Daily: []forecast.DailyForecast{
			{
				Date:             retrievedAt.Format(time.DateOnly),
				Temperature2MMax: &temperature,
			},
		},
	}

	forecastEvent, err := event.NewLatestForecastEventV1(snapshot)
	if err != nil {
		t.Fatalf("create latest forecast event: %v", err)
	}

	return forecastEvent
}
