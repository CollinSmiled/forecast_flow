package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	cloudbigquery "cloud.google.com/go/bigquery"

	"github.com/CollinSmiled/forecast_flow/internal/location"
	forecastbigquery "github.com/CollinSmiled/forecast_flow/internal/platform/bigquery"
	"github.com/CollinSmiled/forecast_flow/internal/platform/postgres"
	"github.com/CollinSmiled/forecast_flow/internal/referencedata"
)

const (
	defaultBigQueryReferenceDataset = "forecast_reference"
	defaultBigQueryLocation         = "asia-southeast2"
)

type config struct {
	DatabaseURL        string
	GoogleCloudProject string
	BigQueryDataset    string
	BigQueryLocation   string
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
		logger.Error("location reference synchronization failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	settings, err := loadConfig()
	if err != nil {
		return err
	}

	database, err := postgres.Open(ctx, settings.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open PostgreSQL: %w", err)
	}
	defer database.Close()

	client, err := cloudbigquery.NewClient(ctx, settings.GoogleCloudProject)
	if err != nil {
		return fmt.Errorf("create BigQuery client: %w", err)
	}
	defer client.Close()

	writer, err := forecastbigquery.NewReferenceWriter(
		client,
		settings.BigQueryDataset,
		settings.BigQueryLocation,
	)
	if err != nil {
		return fmt.Errorf("create location reference writer: %w", err)
	}

	sync, err := referencedata.NewLocationSync(
		location.NewRepository(database),
		writer,
	)
	if err != nil {
		return fmt.Errorf("create location reference sync: %w", err)
	}

	count, err := sync.Run(ctx)
	if err != nil {
		return err
	}

	logger.Info(
		"location reference synchronization completed",
		"location_count", count,
		"project", settings.GoogleCloudProject,
		"dataset", settings.BigQueryDataset,
		"location", settings.BigQueryLocation,
	)

	return nil
}

func loadConfig() (config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	projectID := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		return config{}, errors.New("GOOGLE_CLOUD_PROJECT is required")
	}

	return config{
		DatabaseURL:        databaseURL,
		GoogleCloudProject: projectID,
		BigQueryDataset: envOrDefault(
			"BIGQUERY_REFERENCE_DATASET",
			defaultBigQueryReferenceDataset,
		),
		BigQueryLocation: envOrDefault(
			"BIGQUERY_LOCATION",
			defaultBigQueryLocation,
		),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}
