package event

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

func TestNewVerificationWeatherEventV1(t *testing.T) {
	snapshot := testVerificationSnapshot()

	weatherEvent, err := NewVerificationWeatherEventV1(snapshot)
	if err != nil {
		t.Fatalf("create verification weather event: %v", err)
	}

	if weatherEvent.EventType != VerificationWeatherEventType {
		t.Errorf("event type = %q, want %q", weatherEvent.EventType, VerificationWeatherEventType)
	}
	if weatherEvent.PartitionKey() != "3" {
		t.Errorf("partition key = %q, want 3", weatherEvent.PartitionKey())
	}
	if !weatherEvent.Data.PeriodStart.Equal(snapshot.Hourly[0].ValidAt) {
		t.Errorf("period start = %v, want %v", weatherEvent.Data.PeriodStart, snapshot.Hourly[0].ValidAt)
	}
	if !weatherEvent.Data.PeriodEnd.Equal(snapshot.Hourly[1].ValidAt) {
		t.Errorf("period end = %v, want %v", weatherEvent.Data.PeriodEnd, snapshot.Hourly[1].ValidAt)
	}
}

func TestVerificationWeatherEventIDIsDeterministic(t *testing.T) {
	snapshot := testVerificationSnapshot()

	first, err := NewVerificationWeatherEventV1(snapshot)
	if err != nil {
		t.Fatalf("create first event: %v", err)
	}
	second, err := NewVerificationWeatherEventV1(snapshot)
	if err != nil {
		t.Fatalf("create second event: %v", err)
	}
	if first.EventID != second.EventID {
		t.Errorf("event IDs differ: %q and %q", first.EventID, second.EventID)
	}

	snapshot.RetrievedAt = snapshot.RetrievedAt.Add(time.Hour)
	third, err := NewVerificationWeatherEventV1(snapshot)
	if err != nil {
		t.Fatalf("create third event: %v", err)
	}
	if first.EventID == third.EventID {
		t.Fatal("event ID did not change with retrieval time")
	}
}

func TestVerificationWeatherEventUsesStableJSONNames(t *testing.T) {
	weatherEvent, err := NewVerificationWeatherEventV1(testVerificationSnapshot())
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	body, err := json.Marshal(weatherEvent)
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}
	encoded := string(body)

	for _, field := range []string{
		`"reference_kind":"reanalysis"`,
		`"period_start"`,
		`"period_end"`,
		`"temperature_2m":29.4`,
	} {
		if !strings.Contains(encoded, field) {
			t.Errorf("JSON does not contain %s: %s", field, encoded)
		}
	}
}

func TestVerificationWeatherEventRejectsInvalidSnapshot(t *testing.T) {
	snapshot := testVerificationSnapshot()
	snapshot.Hourly = nil

	if _, err := NewVerificationWeatherEventV1(snapshot); err == nil {
		t.Fatal("expected an error")
	}
}

func testVerificationSnapshot() verification.Snapshot {
	validAt := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC)
	temperature := 29.4

	return verification.Snapshot{
		LocationID:    3,
		Source:        "open_meteo_archive_best_match",
		ReferenceKind: verification.ReferenceKindReanalysis,
		RetrievedAt:   validAt.Add(48 * time.Hour),
		Timezone:      "Asia/Jakarta",
		Hourly: []verification.HourlyWeather{
			{
				ValidAt: validAt,
				WeatherMetrics: forecast.WeatherMetrics{
					Temperature2M: &temperature,
				},
			},
			{ValidAt: validAt.Add(time.Hour)},
		},
	}
}
