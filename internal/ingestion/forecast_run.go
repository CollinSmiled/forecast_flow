package ingestion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type ForecastRunFetcher interface {
	FetchForecastRun(
		ctx context.Context,
		selectedLocation location.Location,
		model forecast.Model,
		forecastRunAt time.Time,
		forecastDays int,
	) (forecast.ForecastRun, error)
}

type ForecastRunPublisher interface {
	PublishForecastRun(
		ctx context.Context,
		forecastEvent event.ForecastRunEventV1,
	) error
}

type ForecastRunIngestor struct {
	fetcher      ForecastRunFetcher
	publisher    ForecastRunPublisher
	forecastDays int
}

func NewForecastRunIngestor(
	fetcher ForecastRunFetcher,
	publisher ForecastRunPublisher,
	forecastDays int,
) (*ForecastRunIngestor, error) {
	if fetcher == nil {
		return nil, errors.New("forecast run fetcher is required")
	}

	if publisher == nil {
		return nil, errors.New("forecast run publisher is required")
	}

	if forecastDays < 1 {
		return nil, errors.New("forecast days must be greater than zero")
	}

	return &ForecastRunIngestor{
		fetcher:      fetcher,
		publisher:    publisher,
		forecastDays: forecastDays,
	}, nil
}

func (ingestor *ForecastRunIngestor) Ingest(
	ctx context.Context,
	selectedLocation location.Location,
	model forecast.Model,
	forecastRunAt time.Time,
) (event.ForecastRunEventV1, error) {
	modelRun, err := ingestor.fetcher.FetchForecastRun(
		ctx,
		selectedLocation,
		model,
		forecastRunAt,
		ingestor.forecastDays,
	)
	if err != nil {
		return event.ForecastRunEventV1{}, fmt.Errorf(
			"fetch forecast run: %w",
			err,
		)
	}

	forecastEvent, err := event.NewForecastRunEventV1(modelRun)
	if err != nil {
		return event.ForecastRunEventV1{}, fmt.Errorf(
			"create forecast run event: %w",
			err,
		)
	}

	if err := ingestor.publisher.PublishForecastRun(
		ctx,
		forecastEvent,
	); err != nil {
		return event.ForecastRunEventV1{}, fmt.Errorf(
			"publish forecast run: %w",
			err,
		)
	}

	return forecastEvent, nil
}
