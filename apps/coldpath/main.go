package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	cloudbigquery "cloud.google.com/go/bigquery"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
	forecastbigquery "github.com/CollinSmiled/forecast_flow/internal/platform/bigquery"
	forecastkafka "github.com/CollinSmiled/forecast_flow/internal/platform/kafka"
)

const (
	defaultKafkaBrokers          = "localhost:9092"
	defaultKafkaConsumerGroup    = "bigquery-cold-path"
	defaultBigQueryDataset       = "forecast_raw"
	defaultBigQueryLocation      = "asia-southeast2"
	defaultColdPathBatchInterval = time.Hour
	defaultColdPathMaxBatchSize  = 500
)

type config struct {
	KafkaBrokers       []string
	KafkaConsumerGroup string
	GoogleCloudProject string
	BigQueryDataset    string
	BigQueryLocation   string
	BatchInterval      time.Duration
	MaxBatchRecords    int
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("cold-path service stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	settings, err := loadConfig()
	if err != nil {
		return err
	}

	client, err := cloudbigquery.NewClient(ctx, settings.GoogleCloudProject)
	if err != nil {
		return fmt.Errorf("create BigQuery client: %w", err)
	}
	defer client.Close()

	writer, err := forecastbigquery.NewWriter(
		client,
		settings.BigQueryDataset,
		settings.BigQueryLocation,
	)
	if err != nil {
		return fmt.Errorf("create BigQuery writer: %w", err)
	}

	processor, err := coldstore.NewProcessor(writer)
	if err != nil {
		return fmt.Errorf("create cold-path processor: %w", err)
	}

	consumer, err := forecastkafka.NewColdPathConsumer(
		ctx,
		forecastkafka.ColdPathConsumerConfig{
			Brokers:         settings.KafkaBrokers,
			GroupID:         settings.KafkaConsumerGroup,
			BatchInterval:   settings.BatchInterval,
			MaxBatchRecords: settings.MaxBatchRecords,
		},
		processor,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create cold-path consumer: %w", err)
	}
	defer consumer.Close()

	logger.Info(
		"cold-path service started",
		"group_id", settings.KafkaConsumerGroup,
		"project", settings.GoogleCloudProject,
		"dataset", settings.BigQueryDataset,
		"location", settings.BigQueryLocation,
		"batch_interval", settings.BatchInterval,
		"max_batch_records", settings.MaxBatchRecords,
	)

	return consumer.Run(ctx)
}

func loadConfig() (config, error) {
	projectID := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		return config{}, errors.New("GOOGLE_CLOUD_PROJECT is required")
	}

	batchInterval, err := positiveDurationOrDefault(
		"COLD_PATH_BATCH_INTERVAL",
		defaultColdPathBatchInterval,
	)
	if err != nil {
		return config{}, err
	}

	maxBatchRecords, err := positiveIntOrDefault(
		"COLD_PATH_MAX_BATCH_RECORDS",
		defaultColdPathMaxBatchSize,
	)
	if err != nil {
		return config{}, err
	}

	return config{
		KafkaBrokers: strings.Split(
			envOrDefault("KAFKA_BROKERS", defaultKafkaBrokers),
			",",
		),
		KafkaConsumerGroup: envOrDefault(
			"KAFKA_COLD_PATH_CONSUMER_GROUP",
			defaultKafkaConsumerGroup,
		),
		GoogleCloudProject: projectID,
		BigQueryDataset: envOrDefault(
			"BIGQUERY_DATASET",
			defaultBigQueryDataset,
		),
		BigQueryLocation: envOrDefault(
			"BIGQUERY_LOCATION",
			defaultBigQueryLocation,
		),
		BatchInterval:   batchInterval,
		MaxBatchRecords: maxBatchRecords,
	}, nil
}

func positiveDurationOrDefault(
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}

	return parsed, nil
}

func positiveIntOrDefault(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return parsed, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}
