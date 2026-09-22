package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestVerificationWeatherBatchUsesEachLocationsLocalDate(t *testing.T) {
	referenceTime := time.Date(2026, time.September, 22, 16, 0, 0, 0, time.UTC)
	locations := &fakeVerificationLocationSource{locations: []location.Location{
		{ID: 3, City: "Jakarta", Timezone: "Asia/Jakarta"},
		{ID: 4, City: "Tokyo", Timezone: "Asia/Tokyo"},
	}}
	runner := &fakeVerificationWeatherRunner{}
	batch, err := NewVerificationWeatherBatch(locations, runner, 7)
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	result, err := batch.Run(context.Background(), referenceTime)
	if err != nil {
		t.Fatalf("run batch: %v", err)
	}

	if result.Locations != 2 || len(result.Published) != 2 {
		t.Fatalf("result = %#v, want two published locations", result)
	}
	if got := runner.calls[0].startDate.Format(time.DateOnly); got != "2026-09-15" {
		t.Errorf("Jakarta target date = %s, want 2026-09-15", got)
	}
	if got := runner.calls[1].startDate.Format(time.DateOnly); got != "2026-09-16" {
		t.Errorf("Tokyo target date = %s, want 2026-09-16", got)
	}
	for _, call := range runner.calls {
		if !call.startDate.Equal(call.endDate) {
			t.Errorf("date range = %v to %v, want one day", call.startDate, call.endDate)
		}
	}
}

func TestVerificationWeatherBatchContinuesAfterLocationFailure(t *testing.T) {
	locations := &fakeVerificationLocationSource{locations: []location.Location{
		{ID: 3, Timezone: "Invalid/Timezone"},
		{ID: 4, Timezone: "Asia/Jakarta"},
		{ID: 5, Timezone: "Asia/Singapore"},
	}}
	runner := &fakeVerificationWeatherRunner{
		failures: map[int64]error{4: errors.New("provider unavailable")},
	}
	batch, err := NewVerificationWeatherBatch(locations, runner, 7)
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	result, err := batch.Run(
		context.Background(),
		time.Date(2026, time.September, 22, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("run batch: %v", err)
	}

	if len(result.Failures) != 2 {
		t.Fatalf("failures = %d, want 2", len(result.Failures))
	}
	if len(result.Published) != 1 || result.Published[0].Location.ID != 5 {
		t.Fatalf("published = %#v, want location 5", result.Published)
	}
}

func TestVerificationWeatherBatchReturnsLocationQueryFailure(t *testing.T) {
	queryError := errors.New("database unavailable")
	locations := &fakeVerificationLocationSource{err: queryError}
	runner := &fakeVerificationWeatherRunner{}
	batch, err := NewVerificationWeatherBatch(locations, runner, 7)
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	_, err = batch.Run(context.Background(), time.Now())
	if !errors.Is(err, queryError) {
		t.Fatalf("error = %v, want wrapped query error", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("ingestion calls = %d, want 0", len(runner.calls))
	}
}

func TestNewVerificationWeatherBatchValidatesConfiguration(t *testing.T) {
	locations := &fakeVerificationLocationSource{}
	runner := &fakeVerificationWeatherRunner{}

	if _, err := NewVerificationWeatherBatch(nil, runner, 7); err == nil {
		t.Error("expected a missing location source error")
	}
	if _, err := NewVerificationWeatherBatch(locations, nil, 7); err == nil {
		t.Error("expected a missing weather runner error")
	}
	if _, err := NewVerificationWeatherBatch(locations, runner, 0); err == nil {
		t.Error("expected an invalid lag error")
	}
}

type fakeVerificationLocationSource struct {
	locations []location.Location
	err       error
}

func (source *fakeVerificationLocationSource) ListAll(
	context.Context,
) ([]location.Location, error) {
	return source.locations, source.err
}

type verificationWeatherCall struct {
	locationID int64
	startDate  time.Time
	endDate    time.Time
}

type fakeVerificationWeatherRunner struct {
	calls    []verificationWeatherCall
	failures map[int64]error
}

func (runner *fakeVerificationWeatherRunner) Ingest(
	_ context.Context,
	selectedLocation location.Location,
	startDate time.Time,
	endDate time.Time,
) (event.VerificationWeatherEventV1, error) {
	runner.calls = append(runner.calls, verificationWeatherCall{
		locationID: selectedLocation.ID,
		startDate:  startDate,
		endDate:    endDate,
	})
	if err := runner.failures[selectedLocation.ID]; err != nil {
		return event.VerificationWeatherEventV1{}, err
	}

	return event.VerificationWeatherEventV1{EventID: "verification-event"}, nil
}
