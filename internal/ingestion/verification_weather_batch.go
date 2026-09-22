package ingestion

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type VerificationWeatherLocationSource interface {
	ListAll(ctx context.Context) ([]location.Location, error)
}

type VerificationWeatherRunner interface {
	Ingest(
		ctx context.Context,
		selectedLocation location.Location,
		startDate time.Time,
		endDate time.Time,
	) (event.VerificationWeatherEventV1, error)
}

type PublishedVerificationWeather struct {
	Location   location.Location
	TargetDate time.Time
	Event      event.VerificationWeatherEventV1
}

type VerificationWeatherFailure struct {
	Location   location.Location
	TargetDate time.Time
	Err        error
}

type VerificationWeatherBatchResult struct {
	Locations int
	Published []PublishedVerificationWeather
	Failures  []VerificationWeatherFailure
}

type VerificationWeatherBatch struct {
	locations VerificationWeatherLocationSource
	ingestor  VerificationWeatherRunner
	lagDays   int
}

func NewVerificationWeatherBatch(
	locations VerificationWeatherLocationSource,
	ingestor VerificationWeatherRunner,
	lagDays int,
) (*VerificationWeatherBatch, error) {
	if locations == nil {
		return nil, errors.New("verification location source is required")
	}
	if ingestor == nil {
		return nil, errors.New("verification weather runner is required")
	}
	if lagDays < 1 {
		return nil, errors.New("verification lag days must be greater than zero")
	}

	return &VerificationWeatherBatch{
		locations: locations,
		ingestor:  ingestor,
		lagDays:   lagDays,
	}, nil
}

func (batch *VerificationWeatherBatch) Run(
	ctx context.Context,
	referenceTime time.Time,
) (VerificationWeatherBatchResult, error) {
	if referenceTime.IsZero() {
		return VerificationWeatherBatchResult{}, errors.New(
			"verification reference time is required",
		)
	}

	locations, err := batch.locations.ListAll(ctx)
	if err != nil {
		return VerificationWeatherBatchResult{}, fmt.Errorf(
			"list verification locations: %w",
			err,
		)
	}

	result := VerificationWeatherBatchResult{
		Locations: len(locations),
		Published: make([]PublishedVerificationWeather, 0, len(locations)),
		Failures:  make([]VerificationWeatherFailure, 0),
	}

	for _, selectedLocation := range locations {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		targetDate, err := verificationTargetDate(
			referenceTime,
			selectedLocation.Timezone,
			batch.lagDays,
		)
		if err != nil {
			result.Failures = append(result.Failures, VerificationWeatherFailure{
				Location: selectedLocation,
				Err:      err,
			})
			continue
		}

		publishedEvent, err := batch.ingestor.Ingest(
			ctx,
			selectedLocation,
			targetDate,
			targetDate,
		)
		if err != nil {
			result.Failures = append(result.Failures, VerificationWeatherFailure{
				Location:   selectedLocation,
				TargetDate: targetDate,
				Err:        err,
			})
			continue
		}

		result.Published = append(result.Published, PublishedVerificationWeather{
			Location:   selectedLocation,
			TargetDate: targetDate,
			Event:      publishedEvent,
		})
	}

	return result, nil
}

func verificationTargetDate(
	referenceTime time.Time,
	timezone string,
	lagDays int,
) (time.Time, error) {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return time.Time{}, errors.New("location timezone is required")
	}

	zone, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load location timezone %q: %w", timezone, err)
	}

	localDate := referenceTime.In(zone).AddDate(0, 0, -lagDays)
	return time.Date(
		localDate.Year(),
		localDate.Month(),
		localDate.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	), nil
}
