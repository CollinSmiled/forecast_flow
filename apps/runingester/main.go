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

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/ingestion"
	"github.com/CollinSmiled/forecast_flow/internal/location"
	forecastkafka "github.com/CollinSmiled/forecast_flow/internal/platform/kafka"
	"github.com/CollinSmiled/forecast_flow/internal/platform/openmeteo"
	"github.com/CollinSmiled/forecast_flow/internal/platform/postgres"
)

const (
	defaultKafkaBrokers = "localhost:9092"
	defaultForecastDays = 10
)

var ecmwfIFS = forecast.Model{
	ID:           "ecmwf_ifs",
	Name:         "ECMWF IFS HRES",
	Provider:     "ECMWF",
	ResolutionKM: 9,
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
		logger.Error("forecast run ingestion failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	locationID, err := requiredPositiveInt64("LOCATION_ID")
	if err != nil {
		return err
	}

	forecastRunAt, err := requiredRunTime("FORECAST_RUN_AT")
	if err != nil {
		return err
	}

	forecastDays, err := positiveIntOrDefault(
		"FORECAST_DAYS",
		defaultForecastDays,
	)
	if err != nil {
		return err
	}

	database, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	locationRepository := location.NewRepository(database)
	selectedLocation, err := locationRepository.GetByID(ctx, locationID)
	if err != nil {
		return fmt.Errorf("load ingestion location: %w", err)
	}

	publisher, err := forecastkafka.NewPublisher(
		ctx,
		strings.Split(
			envOrDefault("KAFKA_BROKERS", defaultKafkaBrokers),
			",",
		),
	)
	if err != nil {
		return fmt.Errorf("create Kafka publisher: %w", err)
	}
	defer publisher.Close()

	ingestor, err := ingestion.NewForecastRunIngestor(
		openmeteo.NewSingleRunClient(),
		publisher,
		forecastDays,
	)
	if err != nil {
		return fmt.Errorf("create forecast run ingestor: %w", err)
	}

	publishedEvent, err := ingestor.Ingest(
		ctx,
		selectedLocation,
		ecmwfIFS,
		forecastRunAt,
	)
	if err != nil {
		return err
	}

	logger.Info(
		"forecast run published",
		"event_id", publishedEvent.EventID,
		"location_id", selectedLocation.ID,
		"city", selectedLocation.City,
		"model_id", publishedEvent.Data.Model.ModelID,
		"forecast_run_at", publishedEvent.Data.ForecastRunAt,
		"hourly_rows", len(publishedEvent.Data.Hourly),
		"topic", event.ForecastRunTopic,
	)

	return nil
}

func requiredRunTime(key string) (time.Time, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return time.Time{}, fmt.Errorf("%s is required", key)
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"%s must use RFC3339 format: %w",
			key,
			err,
		)
	}

	parsed = parsed.UTC()
	if parsed.Minute() != 0 || parsed.Second() != 0 ||
		parsed.Nanosecond() != 0 {
		return time.Time{}, fmt.Errorf(
			"%s must align to a whole UTC hour",
			key,
		)
	}

	return parsed, nil
}

func requiredPositiveInt64(key string) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
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
