package forecast

import (
	"testing"
	"time"
)

func TestNewHourlyForecastCalculatesLeadTime(t *testing.T) {
	run := Run{
		ForecastRunAt: time.Date(
			2026, time.September, 11, 6, 0, 0, 0, time.UTC,
		),
	}

	validAt := time.Date(
		2026, time.September, 11, 18, 0, 0, 0, time.UTC,
	)

	hourly, err := NewHourlyForecast(run, validAt)
	if err != nil {
		t.Fatalf("create hourly forecast: %v", err)
	}

	if hourly.LeadTimeHours != 12 {
		t.Fatalf(
			"expected lead time 12 hours, got %d",
			hourly.LeadTimeHours,
		)
	}

	if !hourly.ValidAt.Equal(validAt) {
		t.Fatalf(
			"expected valid time %v, got %v",
			validAt,
			hourly.ValidAt,
		)
	}
}

func TestNewHourlyForecastRejectsTimeBeforeRun(t *testing.T) {
	run := Run{
		ForecastRunAt: time.Date(
			2026, time.September, 11, 6, 0, 0, 0, time.UTC,
		),
	}

	validAt := time.Date(
		2026, time.September, 11, 5, 0, 0, 0, time.UTC,
	)

	if _, err := NewHourlyForecast(run, validAt); err == nil {
		t.Fatal("expected an error for a valid time before the run")
	}
}

func TestNewDailyForecast(t *testing.T) {
	daily, err := NewDailyForecast("2026-09-11")
	if err != nil {
		t.Fatalf("create daily forecast: %v", err)
	}

	if daily.Date != "2026-09-11" {
		t.Fatalf("expected date 2026-09-11, got %q", daily.Date)
	}
}

func TestNewDailyForecastRejectsTimestamp(t *testing.T) {
	_, err := NewDailyForecast("2026-09-11T00:00:00Z")
	if err == nil {
		t.Fatal("expected an error for a timestamp used as a date")
	}
}

func TestNewDailyForecastRejectsInvalidDate(t *testing.T) {
	_, err := NewDailyForecast("2026-02-30")
	if err == nil {
		t.Fatal("expected an error for an invalid calendar date")
	}
}
