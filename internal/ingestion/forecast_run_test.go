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

func TestForecastRunIngestor(t *testing.T) {
	forecastRun := validForecastRun(t)
	fetcher := &fakeForecastRunFetcher{result: forecastRun}
	publisher := &fakeForecastRunPublisher{}
	ingestor, err := NewForecastRunIngestor(fetcher, publisher, 10)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	selectedLocation := location.Location{ID: 3, City: "Jakarta"}
	publishedEvent, err := ingestor.Ingest(
		context.Background(),
		selectedLocation,
		forecastRun.Run.Model,
		forecastRun.Run.ForecastRunAt,
	)
	if err != nil {
		t.Fatalf("ingest forecast run: %v", err)
	}

	if fetcher.location.ID != 3 {
		t.Errorf("location ID = %d, want 3", fetcher.location.ID)
	}

	if fetcher.forecastDays != 10 {
		t.Errorf("forecast days = %d, want 10", fetcher.forecastDays)
	}

	if publisher.calls != 1 {
		t.Errorf("publish calls = %d, want 1", publisher.calls)
	}

	if publisher.forecastEvent.EventID != publishedEvent.EventID {
		t.Errorf(
			"published event ID = %q, returned event ID = %q",
			publisher.forecastEvent.EventID,
			publishedEvent.EventID,
		)
	}
}

func TestForecastRunIngestorStopsWhenFetchFails(t *testing.T) {
	fetchError := errors.New("provider unavailable")
	fetcher := &fakeForecastRunFetcher{err: fetchError}
	publisher := &fakeForecastRunPublisher{}
	ingestor, err := NewForecastRunIngestor(fetcher, publisher, 10)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	_, err = ingestor.Ingest(
		context.Background(),
		location.Location{ID: 3},
		forecast.Model{},
		time.Now(),
	)
	if !errors.Is(err, fetchError) {
		t.Fatalf("error = %v, want wrapped fetch error", err)
	}

	if publisher.calls != 0 {
		t.Errorf("publish calls = %d, want 0", publisher.calls)
	}
}

func TestForecastRunIngestorReturnsPublishFailure(t *testing.T) {
	publishError := errors.New("broker unavailable")
	forecastRun := validForecastRun(t)
	fetcher := &fakeForecastRunFetcher{result: forecastRun}
	publisher := &fakeForecastRunPublisher{err: publishError}
	ingestor, err := NewForecastRunIngestor(fetcher, publisher, 10)
	if err != nil {
		t.Fatalf("create ingestor: %v", err)
	}

	_, err = ingestor.Ingest(
		context.Background(),
		location.Location{ID: 3},
		forecastRun.Run.Model,
		forecastRun.Run.ForecastRunAt,
	)
	if !errors.Is(err, publishError) {
		t.Fatalf("error = %v, want wrapped publish error", err)
	}
}

type fakeForecastRunFetcher struct {
	result        forecast.ForecastRun
	err           error
	location      location.Location
	model         forecast.Model
	forecastRunAt time.Time
	forecastDays  int
}

func (fetcher *fakeForecastRunFetcher) FetchForecastRun(
	_ context.Context,
	selectedLocation location.Location,
	model forecast.Model,
	forecastRunAt time.Time,
	forecastDays int,
) (forecast.ForecastRun, error) {
	fetcher.location = selectedLocation
	fetcher.model = model
	fetcher.forecastRunAt = forecastRunAt
	fetcher.forecastDays = forecastDays

	return fetcher.result, fetcher.err
}

type fakeForecastRunPublisher struct {
	forecastEvent event.ForecastRunEventV1
	err           error
	calls         int
}

func (publisher *fakeForecastRunPublisher) PublishForecastRun(
	_ context.Context,
	forecastEvent event.ForecastRunEventV1,
) error {
	publisher.calls++
	publisher.forecastEvent = forecastEvent

	return publisher.err
}

func validForecastRun(t *testing.T) forecast.ForecastRun {
	t.Helper()

	runAt := time.Date(
		2026,
		time.September,
		19,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	run, err := forecast.NewRun(
		3,
		forecast.Model{
			ID:           "ecmwf_ifs",
			Name:         "ECMWF IFS HRES",
			Provider:     "ECMWF",
			ResolutionKM: 9,
		},
		runAt,
		runAt.Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	hourly, err := forecast.NewHourlyForecast(
		run,
		runAt.Add(12*time.Hour),
	)
	if err != nil {
		t.Fatalf("create hourly forecast: %v", err)
	}

	return forecast.ForecastRun{
		Run:    run,
		Hourly: []forecast.HourlyForecast{hourly},
	}
}
