package main

import (
	"context"
	"testing"
	"time"
)

func TestRunScheduleRunsImmediatelyAndRepeats(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	err := runSchedule(ctx, time.Millisecond, func(context.Context) {
		calls++
		if calls == 2 {
			cancel()
		}
	})
	if err != nil {
		t.Fatalf("run schedule: %v", err)
	}
	if calls != 2 {
		t.Errorf("cycle calls = %d, want 2", calls)
	}
}

func TestRunScheduleReturnsWithoutRunningWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	if err := runSchedule(ctx, time.Hour, func(context.Context) {
		calls++
	}); err != nil {
		t.Fatalf("run schedule: %v", err)
	}
	if calls != 0 {
		t.Errorf("cycle calls = %d, want 0", calls)
	}
}

func TestRunScheduleValidatesConfiguration(t *testing.T) {
	if err := runSchedule(
		context.Background(),
		0,
		func(context.Context) {},
	); err == nil {
		t.Error("expected invalid interval error")
	}

	if err := runSchedule(
		context.Background(),
		time.Minute,
		nil,
	); err == nil {
		t.Error("expected missing cycle error")
	}
}

func TestPositiveDurationOrDefault(t *testing.T) {
	t.Setenv("FORECAST_POLL_INTERVAL", "")

	value, err := positiveDurationOrDefault(
		"FORECAST_POLL_INTERVAL",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("parse default duration: %v", err)
	}
	if value != time.Minute {
		t.Errorf("duration = %v, want %v", value, time.Minute)
	}

	t.Setenv("FORECAST_POLL_INTERVAL", "30s")
	value, err = positiveDurationOrDefault(
		"FORECAST_POLL_INTERVAL",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("parse duration: %v", err)
	}
	if value != 30*time.Second {
		t.Errorf("duration = %v, want %v", value, 30*time.Second)
	}
}

func TestPositiveDurationOrDefaultRejectsInvalidValue(t *testing.T) {
	t.Setenv("FORECAST_MAX_AGE", "soon")

	if _, err := positiveDurationOrDefault(
		"FORECAST_MAX_AGE",
		time.Hour,
	); err == nil {
		t.Fatal("expected an error")
	}
}

func TestPositiveIntOrDefault(t *testing.T) {
	t.Setenv("FORECAST_BATCH_SIZE", "25")

	value, err := positiveIntOrDefault("FORECAST_BATCH_SIZE", 100)
	if err != nil {
		t.Fatalf("parse batch size: %v", err)
	}
	if value != 25 {
		t.Errorf("batch size = %d, want 25", value)
	}
}
