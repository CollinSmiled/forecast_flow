package main

import "testing"

func TestLoadConfigUsesLocationSyncDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://forecast:test@localhost/forecast")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "forecast-flow-dev")
	t.Setenv("BIGQUERY_REFERENCE_DATASET", "")
	t.Setenv("BIGQUERY_LOCATION", "")

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
}

func TestLoadConfigReadsLocationSyncOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://forecast:test@localhost/forecast")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "weather-analytics")
	t.Setenv("BIGQUERY_REFERENCE_DATASET", "weather_reference")
	t.Setenv("BIGQUERY_LOCATION", "US")

	settings, err := loadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if settings.GoogleCloudProject != "weather-analytics" ||
		settings.BigQueryDataset != "weather_reference" ||
		settings.BigQueryLocation != "US" {
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
