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

	"github.com/CollinSmiled/forecast_flow/internal/hotforecast"
	forecastkafka "github.com/CollinSmiled/forecast_flow/internal/platform/kafka"
	"github.com/CollinSmiled/forecast_flow/internal/platform/postgres"
)

const (
	defaultKafkaBrokers       = "localhost:9092"
	defaultKafkaConsumerGroup = "postgres-hot-path"
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
		logger.Error("hot-path consumer stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	database, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	store := hotforecast.NewRepository(database)
	brokers := strings.Split(
		envOrDefault("KAFKA_BROKERS", defaultKafkaBrokers),
		",",
	)
	groupID := envOrDefault(
		"KAFKA_CONSUMER_GROUP",
		defaultKafkaConsumerGroup,
	)

	consumer, err := forecastkafka.NewLatestForecastConsumer(
		ctx,
		brokers,
		groupID,
		store,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create hot-path consumer: %w", err)
	}
	defer consumer.Close()

	logger.Info(
		"hot-path consumer started",
		"group_id", groupID,
	)

	return consumer.Run(ctx)
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}
