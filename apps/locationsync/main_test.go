package main

import (
	"context"
	"testing"
	"time"
)

func TestLoadConfigUsesLocationSyncDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://forecast:test@localhost/forecast")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "forecast-flow-dev")
	t.Setenv("BIGQUERY_REFERENCE_DATASET", "")
	t.Setenv("BIGQUERY_LOCATION", "")
	t.Setenv("LOCATION_SYNC_INTERVAL", "")

	settings, err := loadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if settings.BigQueryDataset != defaultBigQueryReferenceDataset {
		t.Errorf("dataset = %q, want %q", settings.BigQueryDataset, defaultBigQueryReferenceDataset)
	}
	if settings.BigQueryLocation != defaultBigQueryLocation {
		t.Errorf("location = %q, want %q", settings.BigQueryLocation, defaultBigQueryLocation)
	}
	if settings.LocationSyncInterval != 0 {
		t.Errorf("sync interval = %v, want one-shot mode", settings.LocationSyncInterval)
	}
}

func TestLoadConfigReadsLocationSyncOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://forecast:test@localhost/forecast")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "weather-analytics")
	t.Setenv("BIGQUERY_REFERENCE_DATASET", "weather_reference")
	t.Setenv("BIGQUERY_LOCATION", "US")
	t.Setenv("LOCATION_SYNC_INTERVAL", "30m")

	settings, err := loadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if settings.GoogleCloudProject != "weather-analytics" ||
		settings.BigQueryDataset != "weather_reference" ||
		settings.BigQueryLocation != "US" ||
		settings.LocationSyncInterval != 30*time.Minute {
		t.Fatalf("settings = %+v, want overrides", settings)
	}
}

func TestLoadConfigRequiresLocationSyncConnections(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "forecast-flow-dev")
	if _, err := loadConfig(); err == nil {
		t.Fatal("missing DATABASE_URL error = nil, want an error")
	}

	t.Setenv("DATABASE_URL", "postgres://forecast:test@localhost/forecast")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	if _, err := loadConfig(); err == nil {
		t.Fatal("missing GOOGLE_CLOUD_PROJECT error = nil, want an error")
	}
}

func TestLoadConfigRejectsInvalidSyncInterval(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://forecast:test@localhost/forecast")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "forecast-flow-dev")
	t.Setenv("LOCATION_SYNC_INTERVAL", "immediately")

	if _, err := loadConfig(); err == nil {
		t.Fatal("invalid LOCATION_SYNC_INTERVAL error = nil, want an error")
	}
}

func TestRunScheduleRunsImmediatelyAndRepeats(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	err := runSchedule(ctx, time.Millisecond, func(context.Context) {
		calls++
		if calls == 2 {
			cancel()
		}
	})
	if err != nil {
		t.Fatalf("run schedule: %v", err)
	}
	if calls != 2 {
		t.Errorf("cycle calls = %d, want 2", calls)
	}
}

func TestRunScheduleReturnsWithoutRunningWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	if err := runSchedule(ctx, time.Hour, func(context.Context) {
		calls++
	}); err != nil {
		t.Fatalf("run schedule: %v", err)
	}
	if calls != 0 {
		t.Errorf("cycle calls = %d, want 0", calls)
	}
}

func TestRunScheduleValidatesConfiguration(t *testing.T) {
	if err := runSchedule(context.Background(), 0, func(context.Context) {}); err == nil {
		t.Error("expected invalid interval error")
	}

	if err := runSchedule(context.Background(), time.Hour, nil); err == nil {
		t.Error("expected missing cycle error")
	}
}
