package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/CollinSmiled/forecast_flow/internal/ingestion"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type stubVerificationBatch struct {
	result        ingestion.VerificationWeatherBatchResult
	err           error
	referenceTime time.Time
}

func (batch *stubVerificationBatch) Run(
	_ context.Context,
	referenceTime time.Time,
) (ingestion.VerificationWeatherBatchResult, error) {
	batch.referenceTime = referenceTime
	return batch.result, batch.err
}

func TestRunCycleLogsPublishedAndFailedLocations(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	referenceTime := time.Date(2026, time.September, 22, 0, 0, 0, 0, time.UTC)
	targetDate := referenceTime.AddDate(0, 0, -7)
	batch := &stubVerificationBatch{result: ingestion.VerificationWeatherBatchResult{
		Locations: 2,
		Published: []ingestion.PublishedVerificationWeather{{
			Location:   location.Location{ID: 3, City: "Jakarta"},
			TargetDate: targetDate,
			Event: event.VerificationWeatherEventV1{
				EventID: "verification-event",
				Data: event.VerificationWeatherDataV1{
					Hourly: make([]event.VerificationHourlyWeatherV1, 24),
				},
			},
		}},
		Failures: []ingestion.VerificationWeatherFailure{{
			Location:   location.Location{ID: 4, City: "Tokyo"},
			TargetDate: targetDate,
			Err:        errors.New("provider unavailable"),
		}},
	}}

	runCycle(context.Background(), logger, batch, referenceTime)

	logged := output.String()
	for _, expected := range []string{
		`"msg":"verification weather published"`,
		`"msg":"verification weather ingestion failed"`,
		`"msg":"verification ingestion cycle completed"`,
		`"published":1`,
		`"failed":1`,
	} {
		if !strings.Contains(logged, expected) {
			t.Errorf("logs do not contain %q: %s", expected, logged)
		}
	}
	if !batch.referenceTime.Equal(referenceTime) {
		t.Errorf("reference time = %v, want %v", batch.referenceTime, referenceTime)
	}
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

func TestSchedulerConfigurationParsing(t *testing.T) {
	t.Setenv("VERIFICATION_POLL_INTERVAL", "12h")
	t.Setenv("VERIFICATION_LAG_DAYS", "8")

	interval, err := positiveDurationOrDefault(
		"VERIFICATION_POLL_INTERVAL",
		24*time.Hour,
	)
	if err != nil || interval != 12*time.Hour {
		t.Fatalf("interval = %v, error = %v", interval, err)
	}

	lagDays, err := positiveIntOrDefault("VERIFICATION_LAG_DAYS", 7)
	if err != nil || lagDays != 8 {
		t.Fatalf("lag days = %d, error = %v", lagDays, err)
	}
}
