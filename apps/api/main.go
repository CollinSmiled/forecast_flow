package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/api/httpapi"
	"github.com/CollinSmiled/forecast_flow/internal/hotforecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/CollinSmiled/forecast_flow/internal/platform/openmeteo"
	"github.com/CollinSmiled/forecast_flow/internal/platform/postgres"
)

const (
	defaultHTTPPort            = "8080"
	defaultAllowedCountryCodes = "ID,SG,MY,TH,VN,PH,JP,KR,CN"
	shutdownTimeout            = 10 * time.Second
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
		logger.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	port := envOrDefault("HTTP_PORT", defaultHTTPPort)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	database, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	locationRepository := location.NewRepository(database)
	latestForecastRepository := hotforecast.NewRepository(database)
	geocodingClient := openmeteo.NewGeocodingClient()

	allowedCountryCodes := strings.Split(
		envOrDefault(
			"ALLOWED_COUNTRY_CODES",
			defaultAllowedCountryCodes,
		),
		",",
	)

	locationService, err := location.NewService(
		geocodingClient,
		locationRepository,
		allowedCountryCodes,
	)
	if err != nil {
		return fmt.Errorf(
			"create location service: %w",
			err,
		)
	}

	server := &http.Server{
		Addr: ":" + port,
		Handler: httpapi.NewRouter(
			database,
			locationService,
			latestForecastRepository,
		),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		logger.Info("API listening", "port", port)
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("listen: %w", err)

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}

		err := <-serverError
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("stop server: %w", err)
		}

		return nil
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
