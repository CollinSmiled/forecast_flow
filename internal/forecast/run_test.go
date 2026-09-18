package forecast

import (
	"testing"
	"time"
)

func TestNewRun(t *testing.T) {
	model := Model{
		ID:           "ecmwf_ifs",
		Name:         "ECMWF IFS",
		Provider:     "ECMWF",
		ResolutionKM: 9,
	}

	forecastRunAt := time.Date(
		2026, time.September, 11, 6, 0, 0, 0, time.UTC,
	)
	retrievedAt := time.Date(
		2026, time.September, 11, 10, 0, 0, 0, time.UTC,
	)

	run, err := NewRun(3, model, forecastRunAt, retrievedAt)
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	if run.LocationID != 3 {
		t.Fatalf("expected location ID 3, got %d", run.LocationID)
	}

	if run.Model.ID != "ecmwf_ifs" {
		t.Fatalf("expected ECMWF model, got %q", run.Model.ID)
	}
}

func TestRunLeadTimeHours(t *testing.T) {
	run := Run{
		ForecastRunAt: time.Date(
			2026, time.September, 11, 6, 0, 0, 0, time.UTC,
		),
	}

	validAt := time.Date(
		2026, time.September, 11, 18, 0, 0, 0, time.UTC,
	)

	leadTimeHours, err := run.LeadTimeHours(validAt)
	if err != nil {
		t.Fatalf("calculate lead time: %v", err)
	}

	if leadTimeHours != 12 {
		t.Fatalf("expected 12 lead-time hours, got %d", leadTimeHours)
	}
}

func TestRunRejectsValidTimeBeforeRun(t *testing.T) {
	run := Run{
		ForecastRunAt: time.Date(
			2026, time.September, 11, 6, 0, 0, 0, time.UTC,
		),
	}

	validAt := time.Date(
		2026, time.September, 11, 5, 0, 0, 0, time.UTC,
	)

	if _, err := run.LeadTimeHours(validAt); err == nil {
		t.Fatal("expected an error for a valid time before the run")
	}
}

func TestRunRejectsPartialHourLeadTime(t *testing.T) {
	run := Run{
		ForecastRunAt: time.Date(
			2026, time.September, 11, 6, 0, 0, 0, time.UTC,
		),
	}

	validAt := time.Date(
		2026, time.September, 11, 6, 30, 0, 0, time.UTC,
	)

	if _, err := run.LeadTimeHours(validAt); err == nil {
		t.Fatal("expected an error for a partial-hour lead time")
	}
}
