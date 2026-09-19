package main

import (
	"testing"
	"time"
)

func TestRequiredRunTime(t *testing.T) {
	t.Setenv("FORECAST_RUN_AT", "2026-09-19T00:00:00Z")

	parsed, err := requiredRunTime("FORECAST_RUN_AT")
	if err != nil {
		t.Fatalf("parse run time: %v", err)
	}

	want := time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(want) {
		t.Errorf("run time = %v, want %v", parsed, want)
	}
}

func TestRequiredRunTimeRejectsPartialHour(t *testing.T) {
	t.Setenv("FORECAST_RUN_AT", "2026-09-19T00:30:00Z")

	if _, err := requiredRunTime("FORECAST_RUN_AT"); err == nil {
		t.Fatal("expected an error")
	}
}
