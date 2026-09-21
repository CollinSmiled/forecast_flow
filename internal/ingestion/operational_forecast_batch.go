package ingestion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type DueForecastLocationSource interface {
	ListDueForForecast(
		ctx context.Context,
		staleBefore time.Time,
		limit int,
	) ([]location.Location, error)
}

type OperationalForecastRunner interface {
	Ingest(
		ctx context.Context,
		selectedLocation location.Location,
	) (event.LatestForecastEventV1, error)
}

type PublishedOperationalForecast struct {
	Location location.Location
	Event    event.LatestForecastEventV1
}

type OperationalForecastFailure struct {
	Location location.Location
	Err      error
}

type OperationalForecastBatchResult struct {
	Due       int
	Published []PublishedOperationalForecast
	Failures  []OperationalForecastFailure
}

type OperationalForecastBatch struct {
	locations DueForecastLocationSource
	ingestor  OperationalForecastRunner
	limit     int
}

func NewOperationalForecastBatch(
	locations DueForecastLocationSource,
	ingestor OperationalForecastRunner,
	limit int,
) (*OperationalForecastBatch, error) {
	if locations == nil {
		return nil, errors.New("due forecast location source is required")
	}

	if ingestor == nil {
		return nil, errors.New("operational forecast runner is required")
	}

	if limit < 1 {
		return nil, errors.New("forecast batch limit must be greater than zero")
	}

	return &OperationalForecastBatch{
		locations: locations,
		ingestor:  ingestor,
		limit:     limit,
	}, nil
}

func (batch *OperationalForecastBatch) Run(
	ctx context.Context,
	staleBefore time.Time,
) (OperationalForecastBatchResult, error) {
	locations, err := batch.locations.ListDueForForecast(
		ctx,
		staleBefore,
		batch.limit,
	)
	if err != nil {
		return OperationalForecastBatchResult{}, fmt.Errorf(
			"list due forecast locations: %w",
			err,
		)
	}

	result := OperationalForecastBatchResult{
		Due:       len(locations),
		Published: make([]PublishedOperationalForecast, 0, len(locations)),
		Failures:  make([]OperationalForecastFailure, 0),
	}

	for _, selectedLocation := range locations {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		publishedEvent, err := batch.ingestor.Ingest(ctx, selectedLocation)
		if err != nil {
			result.Failures = append(
				result.Failures,
				OperationalForecastFailure{
					Location: selectedLocation,
					Err:      err,
				},
			)
			continue
		}

		result.Published = append(
			result.Published,
			PublishedOperationalForecast{
				Location: selectedLocation,
				Event:    publishedEvent,
			},
		)
	}

	return result, nil
}
