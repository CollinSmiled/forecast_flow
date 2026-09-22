package coldstore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
)

func TestNewOperationalForecastRow(t *testing.T) {
	instant := testInstant()
	forecastEvent := event.LatestForecastEventV1{
		EventID:       "operational-123",
		EventType:     event.LatestForecastEventType,
		SchemaVersion: event.LatestForecastSchemaVersion,
		OccurredAt:    instant,
		Data: event.LatestForecastDataV1{
			LocationID:  3,
			Source:      "open_meteo_best_match",
			RetrievedAt: instant,
		},
	}
	payload := encodeEvent(t, forecastEvent)

	row, err := NewOperationalForecastRow(
		forecastEvent,
		KafkaRecordMetadata{
			Topic:     event.LatestForecastTopic,
			Partition: 2,
			Offset:    10,
			Key:       "3",
			Timestamp: instant,
		},
		payload,
		instant.Add(time.Second),
	)
	if err != nil {
		t.Fatalf("create operational row: %v", err)
	}

	if row.LocationID != 3 || row.KafkaOffset != 10 {
		t.Errorf(
			"location/offset = %d/%d, want 3/10",
			row.LocationID,
			row.KafkaOffset,
		)
	}

	payload[0] = 'x'
	if row.Payload[0] == 'x' {
		t.Fatal("row retained mutable payload storage")
	}
}

func TestNewModelRunRow(t *testing.T) {
	instant := testInstant()
	forecastEvent := event.ForecastRunEventV1{
		EventID:       "run-123",
		EventType:     event.ForecastRunEventType,
		SchemaVersion: event.ForecastRunSchemaVersion,
		OccurredAt:    instant,
		Data: event.ForecastRunDataV1{
			LocationID: 3,
			Model: event.ForecastModelV1{
				ModelID:  "ecmwf_ifs",
				Provider: "ECMWF",
			},
			ForecastRunAt: instant,
			RetrievedAt:   instant.Add(time.Hour),
		},
	}

	row, err := NewModelRunRow(
		forecastEvent,
		KafkaRecordMetadata{
			Topic:     event.ForecastRunTopic,
			Partition: 0,
			Offset:    4,
			Key:       "3:ecmwf_ifs",
			Timestamp: instant.Add(time.Hour),
		},
		encodeEvent(t, forecastEvent),
		instant.Add(2*time.Hour),
	)
	if err != nil {
		t.Fatalf("create model-run row: %v", err)
	}

	if row.ModelID != "ecmwf_ifs" {
		t.Errorf("model ID = %q, want ecmwf_ifs", row.ModelID)
	}

	if !row.ForecastRunAt.Equal(instant) {
		t.Errorf(
			"forecast run time = %v, want %v",
			row.ForecastRunAt,
			instant,
		)
	}
}

func TestNewModelRunRowRejectsMismatchedKafkaKey(t *testing.T) {
	instant := testInstant()
	forecastEvent := event.ForecastRunEventV1{
		EventID:       "run-123",
		EventType:     event.ForecastRunEventType,
		SchemaVersion: event.ForecastRunSchemaVersion,
		Data: event.ForecastRunDataV1{
			LocationID: 3,
			Model: event.ForecastModelV1{
				ModelID:  "ecmwf_ifs",
				Provider: "ECMWF",
			},
			ForecastRunAt: instant,
			RetrievedAt:   instant,
		},
	}

	_, err := NewModelRunRow(
		forecastEvent,
		KafkaRecordMetadata{
			Topic:     event.ForecastRunTopic,
			Partition: 0,
			Offset:    0,
			Key:       "wrong-key",
			Timestamp: instant,
		},
		encodeEvent(t, forecastEvent),
		instant,
	)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestNewVerificationWeatherRow(t *testing.T) {
	instant := testInstant()
	weatherEvent := event.VerificationWeatherEventV1{
		EventID:       "verification-123",
		EventType:     event.VerificationWeatherEventType,
		SchemaVersion: event.VerificationWeatherSchemaVersion,
		OccurredAt:    instant,
		Data: event.VerificationWeatherDataV1{
			LocationID:    3,
			Source:        "open_meteo_archive",
			ReferenceKind: "reanalysis",
			RetrievedAt:   instant,
			Timezone:      "Asia/Jakarta",
			PeriodStart:   instant.Add(-24 * time.Hour),
			PeriodEnd:     instant.Add(-time.Hour),
		},
	}

	row, err := NewVerificationWeatherRow(
		weatherEvent,
		KafkaRecordMetadata{
			Topic:     event.VerificationWeatherTopic,
			Partition: 2,
			Offset:    8,
			Key:       "3",
			Timestamp: instant,
		},
		encodeEvent(t, weatherEvent),
		instant.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf("create verification-weather row: %v", err)
	}

	if row.LocationID != 3 || row.KafkaOffset != 8 {
		t.Errorf(
			"location/offset = %d/%d, want 3/8",
			row.LocationID,
			row.KafkaOffset,
		)
	}
	if row.ReferenceKind != "reanalysis" {
		t.Errorf("reference kind = %q, want reanalysis", row.ReferenceKind)
	}
}

func TestNewVerificationWeatherRowRejectsMismatchedKafkaKey(t *testing.T) {
	instant := testInstant()
	weatherEvent := event.VerificationWeatherEventV1{
		EventID:       "verification-123",
		EventType:     event.VerificationWeatherEventType,
		SchemaVersion: event.VerificationWeatherSchemaVersion,
		Data: event.VerificationWeatherDataV1{
			LocationID:    3,
			Source:        "open_meteo_archive",
			ReferenceKind: "reanalysis",
			RetrievedAt:   instant,
			PeriodStart:   instant.Add(-24 * time.Hour),
			PeriodEnd:     instant.Add(-time.Hour),
		},
	}

	_, err := NewVerificationWeatherRow(
		weatherEvent,
		KafkaRecordMetadata{
			Topic:     event.VerificationWeatherTopic,
			Partition: 0,
			Offset:    0,
			Key:       "wrong-key",
			Timestamp: instant,
		},
		encodeEvent(t, weatherEvent),
		instant,
	)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func encodeEvent(t *testing.T, value any) []byte {
	t.Helper()

	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}

	return payload
}

func testInstant() time.Time {
	return time.Date(
		2026,
		time.September,
		19,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}
