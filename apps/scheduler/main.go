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
	"github.com/CollinSmiled/forecast_flow/internal/ingestion"
	"github.com/CollinSmiled/forecast_flow/internal/location"
	forecastkafka "github.com/CollinSmiled/forecast_flow/internal/platform/kafka"
	"github.com/CollinSmiled/forecast_flow/internal/platform/openmeteo"
	"github.com/CollinSmiled/forecast_flow/internal/platform/postgres"
)

const (
	defaultKafkaBrokers      = "localhost:9092"
	defaultForecastDays      = 10
	defaultPollInterval      = time.Minute
	defaultForecastMaxAge    = time.Hour
	defaultForecastBatchSize = 100
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
		logger.Error("forecast scheduler stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	forecastDays, err := positiveIntOrDefault(
		"FORECAST_DAYS",
		defaultForecastDays,
	)
	if err != nil {
		return err
	}

	pollInterval, err := positiveDurationOrDefault(
		"FORECAST_POLL_INTERVAL",
		defaultPollInterval,
	)
	if err != nil {
		return err
	}

	forecastMaxAge, err := positiveDurationOrDefault(
		"FORECAST_MAX_AGE",
		defaultForecastMaxAge,
	)
	if err != nil {
		return err
	}

	batchSize, err := positiveIntOrDefault(
		"FORECAST_BATCH_SIZE",
		defaultForecastBatchSize,
	)
	if err != nil {
		return err
	}

	database, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

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

	forecastIngestor, err := ingestion.NewOperationalForecastIngestor(
		openmeteo.NewOperationalForecastClient(),
		publisher,
		forecastDays,
	)
	if err != nil {
		return fmt.Errorf("create forecast ingestor: %w", err)
	}

	batch, err := ingestion.NewOperationalForecastBatch(
		location.NewRepository(database),
		forecastIngestor,
		batchSize,
	)
	if err != nil {
		return fmt.Errorf("create forecast batch: %w", err)
	}

	logger.Info(
		"forecast scheduler started",
		"poll_interval", pollInterval,
		"forecast_max_age", forecastMaxAge,
		"batch_size", batchSize,
	)

	return runSchedule(ctx, pollInterval, func(cycleContext context.Context) {
		runCycle(
			cycleContext,
			logger,
			batch,
			time.Now().UTC().Add(-forecastMaxAge),
		)
	})
}

type forecastBatch interface {
	Run(
		ctx context.Context,
		staleBefore time.Time,
	) (ingestion.OperationalForecastBatchResult, error)
}

func runCycle(
	ctx context.Context,
	logger *slog.Logger,
	batch forecastBatch,
	staleBefore time.Time,
) {
	result, err := batch.Run(ctx, staleBefore)
	if err != nil {
		if ctx.Err() == nil {
			logger.Error("forecast ingestion cycle failed", "error", err)
		}
		return
	}

	for _, failure := range result.Failures {
		logger.Error(
			"operational forecast ingestion failed",
			"location_id", failure.Location.ID,
			"city", failure.Location.City,
			"error", failure.Err,
		)
	}

	for _, published := range result.Published {
		logger.Info(
			"operational forecast published",
			"event_id", published.Event.EventID,
			"location_id", published.Location.ID,
			"city", published.Location.City,
			"topic", event.LatestForecastTopic,
			"event_type", published.Event.EventType,
		)
	}

	completionLevel := slog.LevelDebug
	if result.Due > 0 {
		completionLevel = slog.LevelInfo
	}

	logger.Log(
		ctx,
		completionLevel,
		"forecast ingestion cycle completed",
		"due", result.Due,
		"published", len(result.Published),
		"failed", len(result.Failures),
		"stale_before", staleBefore,
	)
}

func runSchedule(
	ctx context.Context,
	interval time.Duration,
	runCycle func(context.Context),
) error {
	if interval <= 0 {
		return errors.New("schedule interval must be greater than zero")
	}
	if runCycle == nil {
		return errors.New("scheduled cycle is required")
	}
	if ctx.Err() != nil {
		return nil
	}

	runCycle(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runCycle(ctx)
		}
	}
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
