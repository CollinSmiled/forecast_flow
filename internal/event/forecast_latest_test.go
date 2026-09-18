package event

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
)

func TestNewLatestForecastEventV1(t *testing.T) {
	snapshot := testOperationalSnapshot()

	event, err := NewLatestForecastEventV1(snapshot)
	if err != nil {
		t.Fatalf("create latest forecast event: %v", err)
	}

	if event.EventType != LatestForecastEventType {
		t.Errorf(
			"event type = %q, want %q",
			event.EventType,
			LatestForecastEventType,
		)
	}

	if event.SchemaVersion != 1 {
		t.Errorf(
			"schema version = %d, want 1",
			event.SchemaVersion,
		)
	}

	if event.PartitionKey() != "3" {
		t.Errorf(
			"partition key = %q, want 3",
			event.PartitionKey(),
		)
	}

	if len(event.Data.Hourly) != 1 {
		t.Fatalf(
			"hourly count = %d, want 1",
			len(event.Data.Hourly),
		)
	}

	if event.Data.Hourly[0].PrecipitationProbability == nil ||
		*event.Data.Hourly[0].PrecipitationProbability != 75 {
		t.Errorf(
			"probability = %v, want 75",
			event.Data.Hourly[0].PrecipitationProbability,
		)
	}
}

func TestLatestForecastEventIDIsDeterministic(t *testing.T) {
	snapshot := testOperationalSnapshot()

	first, err := NewLatestForecastEventV1(snapshot)
	if err != nil {
		t.Fatalf("create first event: %v", err)
	}

	second, err := NewLatestForecastEventV1(snapshot)
	if err != nil {
		t.Fatalf("create second event: %v", err)
	}

	if first.EventID != second.EventID {
		t.Errorf(
			"event IDs differ: %q and %q",
			first.EventID,
			second.EventID,
		)
	}

	snapshot.RetrievedAt = snapshot.RetrievedAt.Add(time.Minute)
	third, err := NewLatestForecastEventV1(snapshot)
	if err != nil {
		t.Fatalf("create third event: %v", err)
	}

	if first.EventID == third.EventID {
		t.Error("event ID did not change with retrieval time")
	}
}

func TestLatestForecastEventUsesStableJSONNames(t *testing.T) {
	event, err := NewLatestForecastEventV1(
		testOperationalSnapshot(),
	)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	encoded := string(body)

	for _, field := range []string{
		`"event_id"`,
		`"schema_version":1`,
		`"location_id":3`,
		`"temperature_2m":30.4`,
		`"precipitation_probability":75`,
	} {
		if !strings.Contains(encoded, field) {
			t.Errorf(
				"JSON does not contain %s: %s",
				field,
				encoded,
			)
		}
	}

	if strings.Contains(encoded, "Temperature2M") {
		t.Errorf("JSON leaked Go field names: %s", encoded)
	}
}

func TestNewLatestForecastEventV1RejectsInvalidSnapshot(
	t *testing.T,
) {
	snapshot := testOperationalSnapshot()
	snapshot.LocationID = 0

	if _, err := NewLatestForecastEventV1(snapshot); err == nil {
		t.Fatal("expected an error")
	}
}

func testOperationalSnapshot() forecast.OperationalForecastSnapshot {
	retrievedAt := time.Date(
		2026,
		time.September,
		18,
		7,
		0,
		0,
		0,
		time.UTC,
	)
	validAt := time.Date(
		2026,
		time.September,
		18,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	return forecast.OperationalForecastSnapshot{
		LocationID:  3,
		Source:      "open_meteo_best_match",
		RetrievedAt: retrievedAt,
		Timezone:    "Asia/Jakarta",
		Current: forecast.CurrentConditions{
			ValidAt:         retrievedAt,
			IntervalSeconds: 900,
			WeatherMetrics: forecast.WeatherMetrics{
				Temperature2M: float64Pointer(30.1),
				IsDay:         boolPointer(true),
			},
		},
		Hourly: []forecast.OperationalHourlyForecast{
			{
				ValidAt: validAt,
				WeatherMetrics: forecast.WeatherMetrics{
					Temperature2M:            float64Pointer(30.4),
					PrecipitationProbability: float64Pointer(75),
				},
			},
		},
		Daily: []forecast.DailyForecast{
			{
				Date:             "2026-09-18",
				Temperature2MMax: float64Pointer(32.2),
				Temperature2MMin: float64Pointer(25.1),
			},
		},
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
