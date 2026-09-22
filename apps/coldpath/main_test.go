package main

import (
	"testing"
	"time"
)

func TestLoadConfigUsesColdPathDefaults(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "forecast-flow-dev")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("KAFKA_COLD_PATH_CONSUMER_GROUP", "")
	t.Setenv("BIGQUERY_DATASET", "")
	t.Setenv("BIGQUERY_LOCATION", "")
	t.Setenv("COLD_PATH_BATCH_INTERVAL", "")
	t.Setenv("COLD_PATH_MAX_BATCH_RECORDS", "")

	settings, err := loadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(settings.KafkaBrokers) != 1 ||
		settings.KafkaBrokers[0] != defaultKafkaBrokers {
		t.Fatalf("Kafka brokers = %v, want default", settings.KafkaBrokers)
	}
	if settings.KafkaConsumerGroup != defaultKafkaConsumerGroup {
		t.Errorf(
			"consumer group = %q, want %q",
			settings.KafkaConsumerGroup,
			defaultKafkaConsumerGroup,
		)
	}
	if settings.BigQueryDataset != defaultBigQueryDataset {
		t.Errorf("dataset = %q, want %q", settings.BigQueryDataset, defaultBigQueryDataset)
	}
	if settings.BigQueryLocation != defaultBigQueryLocation {
		t.Errorf("location = %q, want %q", settings.BigQueryLocation, defaultBigQueryLocation)
	}
	if settings.BatchInterval != time.Hour {
		t.Errorf("batch interval = %v, want 1h", settings.BatchInterval)
	}
	if settings.MaxBatchRecords != defaultColdPathMaxBatchSize {
		t.Errorf(
			"maximum batch records = %d, want %d",
			settings.MaxBatchRecords,
			defaultColdPathMaxBatchSize,
		)
	}
}

func TestLoadConfigReadsOverrides(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "weather-analytics")
	t.Setenv("KAFKA_BROKERS", "kafka-a:9092,kafka-b:9092")
	t.Setenv("KAFKA_COLD_PATH_CONSUMER_GROUP", "warehouse-consumer")
	t.Setenv("BIGQUERY_DATASET", "weather_raw")
	t.Setenv("BIGQUERY_LOCATION", "US")
	t.Setenv("COLD_PATH_BATCH_INTERVAL", "30m")
	t.Setenv("COLD_PATH_MAX_BATCH_RECORDS", "250")

	settings, err := loadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(settings.KafkaBrokers) != 2 ||
		settings.KafkaBrokers[1] != "kafka-b:9092" {
		t.Fatalf("Kafka brokers = %v, want two overrides", settings.KafkaBrokers)
	}
	if settings.KafkaConsumerGroup != "warehouse-consumer" ||
		settings.GoogleCloudProject != "weather-analytics" ||
		settings.BigQueryDataset != "weather_raw" ||
		settings.BigQueryLocation != "US" ||
		settings.BatchInterval != 30*time.Minute ||
		settings.MaxBatchRecords != 250 {
		t.Fatalf("settings = %+v, want overrides", settings)
	}
}

func TestLoadConfigRequiresGoogleCloudProject(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")

	if _, err := loadConfig(); err == nil {
		t.Fatal("loadConfig() error = nil, want an error")
	}
}

func TestLoadConfigRejectsInvalidBatchSettings(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "forecast-flow-dev")

	t.Run("interval", func(t *testing.T) {
		t.Setenv("COLD_PATH_BATCH_INTERVAL", "0s")
		if _, err := loadConfig(); err == nil {
			t.Fatal("loadConfig() error = nil, want an error")
		}
	})

	t.Run("records", func(t *testing.T) {
		t.Setenv("COLD_PATH_BATCH_INTERVAL", "1h")
		t.Setenv("COLD_PATH_MAX_BATCH_RECORDS", "0")
		if _, err := loadConfig(); err == nil {
			t.Fatal("loadConfig() error = nil, want an error")
		}
	})
}
