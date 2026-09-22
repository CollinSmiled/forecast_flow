package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestOperationalForecastBatchContinuesAfterLocationFailure(t *testing.T) {
	staleBefore := time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC)
	locations := &fakeDueForecastLocationSource{
		locations: []location.Location{
			{ID: 3, City: "Jakarta"},
			{ID: 4, City: "Johor Bahru"},
		},
	}
	runner := &fakeOperationalForecastRunner{
		failures: map[int64]error{3: errors.New("provider unavailable")},
	}

	batch, err := NewOperationalForecastBatch(locations, runner, 25)
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	result, err := batch.Run(context.Background(), staleBefore)
	if err != nil {
		t.Fatalf("run batch: %v", err)
	}

	if locations.staleBefore != staleBefore {
		t.Errorf(
			"stale before = %v, want %v",
			locations.staleBefore,
			staleBefore,
		)
	}
	if locations.limit != 25 {
		t.Errorf("limit = %d, want 25", locations.limit)
	}
	if result.Due != 2 {
		t.Errorf("due = %d, want 2", result.Due)
	}
	if len(result.Failures) != 1 || result.Failures[0].Location.ID != 3 {
		t.Fatalf("failures = %#v, want location 3", result.Failures)
	}
	if len(result.Published) != 1 || result.Published[0].Location.ID != 4 {
		t.Fatalf("published = %#v, want location 4", result.Published)
	}
	if len(runner.locationIDs) != 2 {
		t.Fatalf("ingestion calls = %d, want 2", len(runner.locationIDs))
	}
}

func TestOperationalForecastBatchReturnsLocationQueryFailure(t *testing.T) {
	queryError := errors.New("database unavailable")
	locations := &fakeDueForecastLocationSource{err: queryError}
	runner := &fakeOperationalForecastRunner{}

	batch, err := NewOperationalForecastBatch(locations, runner, 10)
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	_, err = batch.Run(context.Background(), time.Now())
	if !errors.Is(err, queryError) {
		t.Fatalf("error = %v, want wrapped query error", err)
	}
	if len(runner.locationIDs) != 0 {
		t.Errorf("ingestion calls = %d, want 0", len(runner.locationIDs))
	}
}

func TestNewOperationalForecastBatchValidatesDependencies(t *testing.T) {
	runner := &fakeOperationalForecastRunner{}
	locations := &fakeDueForecastLocationSource{}

	if _, err := NewOperationalForecastBatch(nil, runner, 10); err == nil {
		t.Error("expected a missing location source error")
	}
	if _, err := NewOperationalForecastBatch(locations, nil, 10); err == nil {
		t.Error("expected a missing forecast runner error")
	}
	if _, err := NewOperationalForecastBatch(locations, runner, 0); err == nil {
		t.Error("expected an invalid batch limit error")
	}
}

type fakeDueForecastLocationSource struct {
	locations   []location.Location
	err         error
	staleBefore time.Time
	limit       int
}

func (source *fakeDueForecastLocationSource) ListDueForForecast(
	_ context.Context,
	staleBefore time.Time,
	limit int,
) ([]location.Location, error) {
	source.staleBefore = staleBefore
	source.limit = limit
	return source.locations, source.err
}

type fakeOperationalForecastRunner struct {
	failures    map[int64]error
	locationIDs []int64
}

func (runner *fakeOperationalForecastRunner) Ingest(
	_ context.Context,
	selectedLocation location.Location,
) (event.LatestForecastEventV1, error) {
	runner.locationIDs = append(runner.locationIDs, selectedLocation.ID)
	if err := runner.failures[selectedLocation.ID]; err != nil {
		return event.LatestForecastEventV1{}, err
	}

	return event.LatestForecastEventV1{
		EventID: "event-for-location",
	}, nil
}
