package ingestion

import (
	"context"
	"errors"
	"fmt"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type OperationalForecastFetcher interface {
	FetchOperationalForecast(
		ctx context.Context,
		selectedLocation location.Location,
		forecastDays int,
	) (forecast.OperationalForecastSnapshot, error)
}

type LatestForecastPublisher interface {
	PublishLatestForecast(
		ctx context.Context,
		forecastEvent event.LatestForecastEventV1,
	) error
}

type OperationalForecastIngestor struct {
	fetcher      OperationalForecastFetcher
	publisher    LatestForecastPublisher
	forecastDays int
}

func NewOperationalForecastIngestor(
	fetcher OperationalForecastFetcher,
	publisher LatestForecastPublisher,
	forecastDays int,
) (*OperationalForecastIngestor, error) {
	if fetcher == nil {
		return nil, errors.New("operational forecast fetcher is required")
	}

	if publisher == nil {
		return nil, errors.New("latest forecast publisher is required")
	}

	if forecastDays < 1 {
		return nil, errors.New("forecast days must be greater than zero")
	}

	return &OperationalForecastIngestor{
		fetcher:      fetcher,
		publisher:    publisher,
		forecastDays: forecastDays,
	}, nil
}

func (ingestor *OperationalForecastIngestor) Ingest(
	ctx context.Context,
	selectedLocation location.Location,
) (event.LatestForecastEventV1, error) {
	snapshot, err := ingestor.fetcher.FetchOperationalForecast(
		ctx,
		selectedLocation,
		ingestor.forecastDays,
	)
	if err != nil {
		return event.LatestForecastEventV1{}, fmt.Errorf(
			"fetch operational forecast: %w",
			err,
		)
	}

	forecastEvent, err := event.NewLatestForecastEventV1(snapshot)
	if err != nil {
		return event.LatestForecastEventV1{}, fmt.Errorf(
			"create latest forecast event: %w",
			err,
		)
	}

	if err := ingestor.publisher.PublishLatestForecast(
		ctx,
		forecastEvent,
	); err != nil {
		return event.LatestForecastEventV1{}, fmt.Errorf(
			"publish latest forecast: %w",
			err,
		)
	}

	return forecastEvent, nil
}
