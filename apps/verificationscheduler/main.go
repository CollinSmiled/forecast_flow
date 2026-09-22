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
	defaultKafkaBrokers             = "localhost:9092"
	defaultVerificationPollInterval = 24 * time.Hour
	defaultVerificationLagDays      = 7
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
		logger.Error("verification scheduler stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	pollInterval, err := positiveDurationOrDefault(
		"VERIFICATION_POLL_INTERVAL",
		defaultVerificationPollInterval,
	)
	if err != nil {
		return err
	}

	lagDays, err := positiveIntOrDefault(
		"VERIFICATION_LAG_DAYS",
		defaultVerificationLagDays,
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

	weatherIngestor, err := ingestion.NewVerificationWeatherIngestor(
		openmeteo.NewVerificationWeatherClient(),
		publisher,
	)
	if err != nil {
		return fmt.Errorf("create verification weather ingestor: %w", err)
	}

	batch, err := ingestion.NewVerificationWeatherBatch(
		location.NewRepository(database),
		weatherIngestor,
		lagDays,
	)
	if err != nil {
		return fmt.Errorf("create verification weather batch: %w", err)
	}

	logger.Info(
		"verification scheduler started",
		"poll_interval", pollInterval,
		"lag_days", lagDays,
	)

	return runSchedule(ctx, pollInterval, func(cycleContext context.Context) {
		runCycle(cycleContext, logger, batch, time.Now().UTC())
	})
}

type verificationBatch interface {
	Run(
		ctx context.Context,
		referenceTime time.Time,
	) (ingestion.VerificationWeatherBatchResult, error)
}

func runCycle(
	ctx context.Context,
	logger *slog.Logger,
	batch verificationBatch,
	referenceTime time.Time,
) {
	result, err := batch.Run(ctx, referenceTime)
	if err != nil {
		if ctx.Err() == nil {
			logger.Error("verification ingestion cycle failed", "error", err)
		}
		return
	}

	for _, failure := range result.Failures {
		logger.Error(
			"verification weather ingestion failed",
			"location_id", failure.Location.ID,
			"city", failure.Location.City,
			"target_date", formatOptionalDate(failure.TargetDate),
			"error", failure.Err,
		)
	}

	for _, published := range result.Published {
		logger.Info(
			"verification weather published",
			"event_id", published.Event.EventID,
			"location_id", published.Location.ID,
			"city", published.Location.City,
			"target_date", published.TargetDate.Format(time.DateOnly),
			"hourly_rows", len(published.Event.Data.Hourly),
			"topic", event.VerificationWeatherTopic,
		)
	}

	logger.Info(
		"verification ingestion cycle completed",
		"locations", result.Locations,
		"published", len(result.Published),
		"failed", len(result.Failures),
		"reference_time", referenceTime,
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

func formatOptionalDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.DateOnly)
}
