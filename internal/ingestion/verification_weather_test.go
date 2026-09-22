package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

func TestVerificationWeatherIngestor(t *testing.T) {
	snapshot := validVerificationSnapshot()
	fetcher := &fakeVerificationWeatherFetcher{snapshot: snapshot}
	publisher := &fakeVerificationWeatherPublisher{}
	ingestor, err := NewVerificationWeatherIngestor(fetcher, publisher)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	startDate := snapshot.Hourly[0].ValidAt
	endDate := snapshot.Hourly[len(snapshot.Hourly)-1].ValidAt
	publishedEvent, err := ingestor.Ingest(
		context.Background(),
		location.Location{ID: 3},
		startDate,
		endDate,
	)
	if err != nil {
		t.Fatalf("ingest verification weather: %v", err)
	}

	if fetcher.location.ID != 3 || fetcher.startDate != startDate || fetcher.endDate != endDate {
		t.Errorf("fetch arguments = location %d, %v to %v", fetcher.location.ID, fetcher.startDate, fetcher.endDate)
	}
	if publisher.calls != 1 {
		t.Errorf("publish calls = %d, want 1", publisher.calls)
	}
	if publisher.weatherEvent.EventID != publishedEvent.EventID {
		t.Errorf("published event ID = %q, want %q", publisher.weatherEvent.EventID, publishedEvent.EventID)
	}
}

func TestVerificationWeatherIngestorStopsWhenFetchFails(t *testing.T) {
	fetchError := errors.New("provider unavailable")
	fetcher := &fakeVerificationWeatherFetcher{err: fetchError}
	publisher := &fakeVerificationWeatherPublisher{}
	ingestor, err := NewVerificationWeatherIngestor(fetcher, publisher)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	_, err = ingestor.Ingest(context.Background(), location.Location{ID: 3}, time.Now(), time.Now())
	if !errors.Is(err, fetchError) {
		t.Fatalf("error = %v, want wrapped fetch error", err)
	}
	if publisher.calls != 0 {
		t.Errorf("publish calls = %d, want 0", publisher.calls)
	}
}

func TestVerificationWeatherIngestorReturnsPublishFailure(t *testing.T) {
	publishError := errors.New("broker unavailable")
	fetcher := &fakeVerificationWeatherFetcher{snapshot: validVerificationSnapshot()}
	publisher := &fakeVerificationWeatherPublisher{err: publishError}
	ingestor, err := NewVerificationWeatherIngestor(fetcher, publisher)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	_, err = ingestor.Ingest(context.Background(), location.Location{ID: 3}, time.Now(), time.Now())
	if !errors.Is(err, publishError) {
		t.Fatalf("error = %v, want wrapped publish error", err)
	}
}

type fakeVerificationWeatherFetcher struct {
	snapshot  verification.Snapshot
	err       error
	location  location.Location
	startDate time.Time
	endDate   time.Time
}

func (fetcher *fakeVerificationWeatherFetcher) FetchVerificationWeather(
	_ context.Context,
	selectedLocation location.Location,
	startDate time.Time,
	endDate time.Time,
) (verification.Snapshot, error) {
	fetcher.location = selectedLocation
	fetcher.startDate = startDate
	fetcher.endDate = endDate
	return fetcher.snapshot, fetcher.err
}

type fakeVerificationWeatherPublisher struct {
	weatherEvent event.VerificationWeatherEventV1
	err          error
	calls        int
}

func (publisher *fakeVerificationWeatherPublisher) PublishVerificationWeather(
	_ context.Context,
	weatherEvent event.VerificationWeatherEventV1,
) error {
	publisher.calls++
	publisher.weatherEvent = weatherEvent
	return publisher.err
}

func validVerificationSnapshot() verification.Snapshot {
	validAt := time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC)
	return verification.Snapshot{
		LocationID:    3,
		Source:        "open_meteo_historical_weather_best_match",
		ReferenceKind: verification.ReferenceKindReanalysis,
		RetrievedAt:   validAt.Add(48 * time.Hour),
		Timezone:      "Asia/Jakarta",
		Hourly: []verification.HourlyWeather{
			{ValidAt: validAt},
			{ValidAt: validAt.Add(time.Hour)},
		},
	}
}
