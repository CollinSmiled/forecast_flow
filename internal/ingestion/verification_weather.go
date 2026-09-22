package ingestion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

type VerificationWeatherFetcher interface {
	FetchVerificationWeather(
		ctx context.Context,
		selectedLocation location.Location,
		startDate time.Time,
		endDate time.Time,
	) (verification.Snapshot, error)
}

type VerificationWeatherPublisher interface {
	PublishVerificationWeather(
		ctx context.Context,
		weatherEvent event.VerificationWeatherEventV1,
	) error
}

type VerificationWeatherIngestor struct {
	fetcher   VerificationWeatherFetcher
	publisher VerificationWeatherPublisher
}

func NewVerificationWeatherIngestor(
	fetcher VerificationWeatherFetcher,
	publisher VerificationWeatherPublisher,
) (*VerificationWeatherIngestor, error) {
	if fetcher == nil {
		return nil, errors.New("verification weather fetcher is required")
	}
	if publisher == nil {
		return nil, errors.New("verification weather publisher is required")
	}

	return &VerificationWeatherIngestor{
		fetcher:   fetcher,
		publisher: publisher,
	}, nil
}

func (ingestor *VerificationWeatherIngestor) Ingest(
	ctx context.Context,
	selectedLocation location.Location,
	startDate time.Time,
	endDate time.Time,
) (event.VerificationWeatherEventV1, error) {
	snapshot, err := ingestor.fetcher.FetchVerificationWeather(
		ctx,
		selectedLocation,
		startDate,
		endDate,
	)
	if err != nil {
		return event.VerificationWeatherEventV1{}, fmt.Errorf(
			"fetch verification weather: %w",
			err,
		)
	}

	weatherEvent, err := event.NewVerificationWeatherEventV1(snapshot)
	if err != nil {
		return event.VerificationWeatherEventV1{}, fmt.Errorf(
			"create verification weather event: %w",
			err,
		)
	}

	if err := ingestor.publisher.PublishVerificationWeather(ctx, weatherEvent); err != nil {
		return event.VerificationWeatherEventV1{}, fmt.Errorf(
			"publish verification weather: %w",
			err,
		)
	}

	return weatherEvent, nil
}
