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

	"github.com/CollinSmiled/forecast_flow/internal/event"
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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("forecast ingestion failed", "error", err)
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

	brokers := strings.Split(
		envOrDefault("KAFKA_BROKERS", defaultKafkaBrokers),
		",",
	)

	publisher, err := forecastkafka.NewPublisher(ctx, brokers)
	if err != nil {
		return fmt.Errorf("create Kafka publisher: %w", err)
	}
	defer publisher.Close()

	ingestor, err := ingestion.NewOperationalForecastIngestor(
		openmeteo.NewOperationalForecastClient(),
		publisher,
		forecastDays,
	)
	if err != nil {
		return fmt.Errorf("create forecast ingestor: %w", err)
	}

	publishedEvent, err := ingestor.Ingest(ctx, selectedLocation)
	if err != nil {
		return err
	}

	logger.Info(
		"operational forecast published",
		"event_id", publishedEvent.EventID,
		"location_id", selectedLocation.ID,
		"city", selectedLocation.City,
		"topic", event.LatestForecastTopic,
		"event_type", publishedEvent.EventType,
	)

	return nil
}

func requiredPositiveInt64(key string) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf(
			"%s must be a positive integer",
			key,
		)
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
		return 0, fmt.Errorf(
			"%s must be a positive integer",
			key,
		)
	}

	return parsed, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}
