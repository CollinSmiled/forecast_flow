package event

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
)

func TestNewForecastRunEventV1(t *testing.T) {
	forecastRun := testForecastRun(t)

	forecastEvent, err := NewForecastRunEventV1(forecastRun)
	if err != nil {
		t.Fatalf("create forecast run event: %v", err)
	}

	if forecastEvent.EventType != ForecastRunEventType {
		t.Errorf(
			"event type = %q, want %q",
			forecastEvent.EventType,
			ForecastRunEventType,
		)
	}

	if forecastEvent.PartitionKey() != "3:ecmwf_ifs" {
		t.Errorf(
			"partition key = %q, want 3:ecmwf_ifs",
			forecastEvent.PartitionKey(),
		)
	}

	if len(forecastEvent.Data.Hourly) != 1 {
		t.Fatalf(
			"hourly count = %d, want 1",
			len(forecastEvent.Data.Hourly),
		)
	}

	if forecastEvent.Data.Hourly[0].LeadTimeHours != 12 {
		t.Errorf(
			"lead time = %d, want 12",
			forecastEvent.Data.Hourly[0].LeadTimeHours,
		)
	}
}

func TestForecastRunEventIDUsesRunIdentity(t *testing.T) {
	forecastRun := testForecastRun(t)

	first, err := NewForecastRunEventV1(forecastRun)
	if err != nil {
		t.Fatalf("create first event: %v", err)
	}

	forecastRun.Run.RetrievedAt = forecastRun.Run.RetrievedAt.Add(time.Hour)
	second, err := NewForecastRunEventV1(forecastRun)
	if err != nil {
		t.Fatalf("create second event: %v", err)
	}

	if first.EventID != second.EventID {
		t.Errorf(
			"same model run produced different event IDs: %q and %q",
			first.EventID,
			second.EventID,
		)
	}

	forecastRun.Run.ForecastRunAt =
		forecastRun.Run.ForecastRunAt.Add(6 * time.Hour)
	forecastRun.Hourly[0].ValidAt =
		forecastRun.Hourly[0].ValidAt.Add(6 * time.Hour)

	third, err := NewForecastRunEventV1(forecastRun)
	if err != nil {
		t.Fatalf("create third event: %v", err)
	}

	if first.EventID == third.EventID {
		t.Fatal("different model runs produced the same event ID")
	}
}

func TestForecastRunEventUsesStableJSONNames(t *testing.T) {
	forecastEvent, err := NewForecastRunEventV1(testForecastRun(t))
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	body, err := json.Marshal(forecastEvent)
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}

	encoded := string(body)
	for _, field := range []string{
		`"forecast_run_at"`,
		`"lead_time_hours":12`,
		`"model_id":"ecmwf_ifs"`,
		`"temperature_2m":30.4`,
	} {
		if !strings.Contains(encoded, field) {
			t.Errorf("JSON does not contain %s: %s", field, encoded)
		}
	}
}

func TestForecastRunEventRejectsIncorrectLeadTime(t *testing.T) {
	forecastRun := testForecastRun(t)
	forecastRun.Hourly[0].LeadTimeHours = 11

	if _, err := NewForecastRunEventV1(forecastRun); err == nil {
		t.Fatal("expected an error")
	}
}

func testForecastRun(t *testing.T) forecast.ForecastRun {
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
	retrievedAt := runAt.Add(30 * time.Minute)
	run, err := forecast.NewRun(
		3,
		forecast.Model{
			ID:           "ecmwf_ifs",
			Name:         "ECMWF IFS",
			Provider:     "ECMWF",
			ResolutionKM: 9,
		},
		runAt,
		retrievedAt,
	)
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	temperature := 30.4
	hourly, err := forecast.NewHourlyForecast(
		run,
		runAt.Add(12*time.Hour),
	)
	if err != nil {
		t.Fatalf("create hourly forecast: %v", err)
	}
	hourly.Temperature2M = &temperature

	return forecast.ForecastRun{
		Run:    run,
		Hourly: []forecast.HourlyForecast{hourly},
	}
}
