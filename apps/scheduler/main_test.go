package main

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/ingestion"
)

type stubForecastBatch struct {
	result ingestion.OperationalForecastBatchResult
	err    error
}

func (batch stubForecastBatch) Run(
	context.Context,
	time.Time,
) (ingestion.OperationalForecastBatchResult, error) {
	return batch.result, batch.err
}

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

func TestRunCycleDoesNotLogEmptyCyclesAtInfo(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	runCycle(
		context.Background(),
		logger,
		stubForecastBatch{},
		time.Now().UTC(),
	)

	if output.Len() != 0 {
		t.Fatalf("empty cycle log = %q, want no info log", output.String())
	}
}

func TestRunCycleLogsCompletedWorkAtInfo(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	runCycle(
		context.Background(),
		logger,
		stubForecastBatch{
			result: ingestion.OperationalForecastBatchResult{Due: 1},
		},
		time.Now().UTC(),
	)

	logged := output.String()
	if !strings.Contains(logged, `"msg":"forecast ingestion cycle completed"`) {
		t.Fatalf("completed cycle log = %q, want completion message", logged)
	}
	if !strings.Contains(logged, `"due":1`) {
		t.Fatalf("completed cycle log = %q, want due count", logged)
	}
}
