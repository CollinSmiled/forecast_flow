package kafka

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestLatestForecastRecord(t *testing.T) {
	occurredAt := time.Date(
		2026,
		time.September,
		18,
		7,
		0,
		0,
		0,
		time.UTC,
	)

	forecastEvent := event.LatestForecastEventV1{
		EventID:       "event-123",
		EventType:     event.LatestForecastEventType,
		SchemaVersion: event.LatestForecastSchemaVersion,
		OccurredAt:    occurredAt,
		Data: event.LatestForecastDataV1{
			LocationID: 3,
		},
	}

	record, err := latestForecastRecord(forecastEvent)
	if err != nil {
		t.Fatalf("build Kafka record: %v", err)
	}

	if record.Topic != event.LatestForecastTopic {
		t.Errorf(
			"topic = %q, want %q",
			record.Topic,
			event.LatestForecastTopic,
		)
	}

	if string(record.Key) != "3" {
		t.Errorf(
			"key = %q, want %q",
			string(record.Key),
			"3",
		)
	}

	if !record.Timestamp.Equal(occurredAt) {
		t.Errorf(
			"timestamp = %v, want %v",
			record.Timestamp,
			occurredAt,
		)
	}

	var decoded event.LatestForecastEventV1

	if err := json.Unmarshal(record.Value, &decoded); err != nil {
		t.Fatalf("decode Kafka record value: %v", err)
	}

	if decoded.EventID != forecastEvent.EventID {
		t.Errorf(
			"event ID = %q, want %q",
			decoded.EventID,
			forecastEvent.EventID,
		)
	}

	headers := recordHeaders(record.Headers)

	if headers["event_type"] != event.LatestForecastEventType {
		t.Errorf(
			"event_type header = %q, want %q",
			headers["event_type"],
			event.LatestForecastEventType,
		)
	}

	if headers["schema_version"] != "1" {
		t.Errorf(
			"schema_version header = %q, want %q",
			headers["schema_version"],
			"1",
		)
	}
}

func TestForecastRunRecord(t *testing.T) {
	occurredAt := time.Date(
		2026,
		time.September,
		19,
		1,
		0,
		0,
		0,
		time.UTC,
	)
	forecastEvent := event.ForecastRunEventV1{
		EventID:       "run-event-123",
		EventType:     event.ForecastRunEventType,
		SchemaVersion: event.ForecastRunSchemaVersion,
		OccurredAt:    occurredAt,
		Data: event.ForecastRunDataV1{
			LocationID: 3,
			Model: event.ForecastModelV1{
				ModelID: "ecmwf_ifs",
			},
		},
	}

	record, err := forecastRunRecord(forecastEvent)
	if err != nil {
		t.Fatalf("build Kafka record: %v", err)
	}

	if record.Topic != event.ForecastRunTopic {
		t.Errorf(
			"topic = %q, want %q",
			record.Topic,
			event.ForecastRunTopic,
		)
	}

	if string(record.Key) != "3:ecmwf_ifs" {
		t.Errorf(
			"key = %q, want 3:ecmwf_ifs",
			string(record.Key),
		)
	}

	if !record.Timestamp.Equal(occurredAt) {
		t.Errorf(
			"timestamp = %v, want %v",
			record.Timestamp,
			occurredAt,
		)
	}

	var decoded event.ForecastRunEventV1
	if err := json.Unmarshal(record.Value, &decoded); err != nil {
		t.Fatalf("decode Kafka record value: %v", err)
	}

	if decoded.EventID != forecastEvent.EventID {
		t.Errorf(
			"event ID = %q, want %q",
			decoded.EventID,
			forecastEvent.EventID,
		)
	}

	headers := recordHeaders(record.Headers)
	if headers["event_type"] != event.ForecastRunEventType {
		t.Errorf(
			"event_type header = %q, want %q",
			headers["event_type"],
			event.ForecastRunEventType,
		)
	}
}

func TestNormalizeBrokers(t *testing.T) {
	brokers := normalizeBrokers([]string{
		" localhost:9092 ",
		"",
		"  ",
		"kafka:19092",
	})

	if len(brokers) != 2 {
		t.Fatalf(
			"broker count = %d, want 2",
			len(brokers),
		)
	}

	if brokers[0] != "localhost:9092" {
		t.Errorf(
			"first broker = %q, want localhost:9092",
			brokers[0],
		)
	}

	if brokers[1] != "kafka:19092" {
		t.Errorf(
			"second broker = %q, want kafka:19092",
			brokers[1],
		)
	}
}

func recordHeaders(
	headers []kgo.RecordHeader,
) map[string]string {
	result := make(map[string]string, len(headers))

	for _, header := range headers {
		result[header.Key] = string(header.Value)
	}

	return result
}
