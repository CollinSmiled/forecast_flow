package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestOperationalForecastIngestor(t *testing.T) {
	selectedLocation := location.Location{
		ID:       3,
		City:     "Jakarta",
		Timezone: "Asia/Jakarta",
	}

	fetcher := &fakeOperationalForecastFetcher{
		snapshot: validOperationalForecastSnapshot(),
	}
	publisher := &fakeLatestForecastPublisher{}

	ingestor, err := NewOperationalForecastIngestor(
		fetcher,
		publisher,
		10,
	)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	publishedEvent, err := ingestor.Ingest(
		context.Background(),
		selectedLocation,
	)
	if err != nil {
		t.Fatalf("ingest forecast: %v", err)
	}

	if fetcher.location.ID != selectedLocation.ID {
		t.Errorf(
			"location ID = %d, want %d",
			fetcher.location.ID,
			selectedLocation.ID,
		)
	}

	if fetcher.forecastDays != 10 {
		t.Errorf(
			"forecast days = %d, want 10",
			fetcher.forecastDays,
		)
	}

	if publisher.calls != 1 {
		t.Fatalf(
			"publish calls = %d, want 1",
			publisher.calls,
		)
	}

	if publisher.forecastEvent.EventID != publishedEvent.EventID {
		t.Errorf(
			"published event ID = %q, returned event ID = %q",
			publisher.forecastEvent.EventID,
			publishedEvent.EventID,
		)
	}

	if publishedEvent.PartitionKey() != "3" {
		t.Errorf(
			"partition key = %q, want 3",
			publishedEvent.PartitionKey(),
		)
	}
}

func TestOperationalForecastIngestorStopsWhenFetchFails(
	t *testing.T,
) {
	fetchError := errors.New("provider unavailable")
	fetcher := &fakeOperationalForecastFetcher{
		err: fetchError,
	}
	publisher := &fakeLatestForecastPublisher{}

	ingestor, err := NewOperationalForecastIngestor(
		fetcher,
		publisher,
		10,
	)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	_, err = ingestor.Ingest(
		context.Background(),
		location.Location{ID: 3},
	)
	if !errors.Is(err, fetchError) {
		t.Fatalf("error = %v, want wrapped fetch error", err)
	}

	if publisher.calls != 0 {
		t.Errorf(
			"publish calls = %d, want 0",
			publisher.calls,
		)
	}
}

func TestOperationalForecastIngestorReturnsPublishFailure(
	t *testing.T,
) {
	publishError := errors.New("broker unavailable")
	fetcher := &fakeOperationalForecastFetcher{
		snapshot: validOperationalForecastSnapshot(),
	}
	publisher := &fakeLatestForecastPublisher{
		err: publishError,
	}

	ingestor, err := NewOperationalForecastIngestor(
		fetcher,
		publisher,
		10,
	)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	_, err = ingestor.Ingest(
		context.Background(),
		location.Location{ID: 3},
	)
	if !errors.Is(err, publishError) {
		t.Fatalf("error = %v, want wrapped publish error", err)
	}
}

type fakeOperationalForecastFetcher struct {
	snapshot     forecast.OperationalForecastSnapshot
	err          error
	location     location.Location
	forecastDays int
}

func (fetcher *fakeOperationalForecastFetcher) FetchOperationalForecast(
	_ context.Context,
	selectedLocation location.Location,
	forecastDays int,
) (forecast.OperationalForecastSnapshot, error) {
	fetcher.location = selectedLocation
	fetcher.forecastDays = forecastDays

	return fetcher.snapshot, fetcher.err
}

type fakeLatestForecastPublisher struct {
	forecastEvent event.LatestForecastEventV1
	err           error
	calls         int
}

func (publisher *fakeLatestForecastPublisher) PublishLatestForecast(
	_ context.Context,
	forecastEvent event.LatestForecastEventV1,
) error {
	publisher.calls++
	publisher.forecastEvent = forecastEvent

	return publisher.err
}

func validOperationalForecastSnapshot() forecast.OperationalForecastSnapshot {
	retrievedAt := time.Date(
		2026,
		time.September,
		18,
		7,
		0,
		0,
		0,
		time.UTC,
	)

	return forecast.OperationalForecastSnapshot{
		LocationID:  3,
		Source:      "open_meteo_best_match",
		RetrievedAt: retrievedAt,
		Timezone:    "Asia/Jakarta",
		Current: forecast.CurrentConditions{
			ValidAt: retrievedAt,
		},
		Hourly: []forecast.OperationalHourlyForecast{
			{
				ValidAt: retrievedAt.Add(time.Hour),
			},
		},
		Daily: []forecast.DailyForecast{
			{
				Date: "2026-09-18",
			},
		},
	}
}
