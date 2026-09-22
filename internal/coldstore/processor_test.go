package coldstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
)

func TestProcessorRoutesOperationalForecast(t *testing.T) {
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
	payload := encodeProcessorEvent(t, forecastEvent)
	writer := &recordingWriter{}
	processor := newTestProcessor(t, writer, instant.Add(time.Minute))

	err := processor.Process(
		context.Background(),
		KafkaRecordMetadata{
			Topic:     event.LatestForecastTopic,
			Partition: 2,
			Offset:    10,
			Key:       forecastEvent.PartitionKey(),
			Timestamp: instant,
		},
		payload,
	)
	if err != nil {
		t.Fatalf("process operational forecast: %v", err)
	}

	if len(writer.operational) != 1 {
		t.Fatalf("operational writes = %d, want 1", len(writer.operational))
	}
	if len(writer.modelRuns) != 0 {
		t.Fatalf("model-run writes = %d, want 0", len(writer.modelRuns))
	}

	stored := writer.operational[0]
	if stored.EventID != forecastEvent.EventID || stored.KafkaOffset != 10 {
		t.Errorf(
			"stored event/offset = %q/%d, want %q/10",
			stored.EventID,
			stored.KafkaOffset,
			forecastEvent.EventID,
		)
	}
	if !stored.IngestedAt.Equal(instant.Add(time.Minute)) {
		t.Errorf("ingested at = %v, want deterministic clock", stored.IngestedAt)
	}
}

func TestProcessorRoutesModelRun(t *testing.T) {
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
	payload := encodeProcessorEvent(t, forecastEvent)
	writer := &recordingWriter{}
	processor := newTestProcessor(t, writer, instant.Add(2*time.Hour))

	err := processor.Process(
		context.Background(),
		KafkaRecordMetadata{
			Topic:     event.ForecastRunTopic,
			Partition: 1,
			Offset:    4,
			Key:       forecastEvent.PartitionKey(),
			Timestamp: instant.Add(time.Hour),
		},
		payload,
	)
	if err != nil {
		t.Fatalf("process model run: %v", err)
	}

	if len(writer.modelRuns) != 1 {
		t.Fatalf("model-run writes = %d, want 1", len(writer.modelRuns))
	}
	if len(writer.operational) != 0 {
		t.Fatalf("operational writes = %d, want 0", len(writer.operational))
	}
	if writer.modelRuns[0].ModelID != "ecmwf_ifs" {
		t.Errorf("model ID = %q, want ecmwf_ifs", writer.modelRuns[0].ModelID)
	}
}

func TestProcessorWritesRecordsAsOneBatchPerTopic(t *testing.T) {
	instant := testInstant()
	writer := &recordingWriter{}
	processor := newTestProcessor(t, writer, instant.Add(time.Minute))
	records := make([]Record, 0, 2)

	for index, eventID := range []string{"operational-1", "operational-2"} {
		forecastEvent := event.LatestForecastEventV1{
			EventID:       eventID,
			EventType:     event.LatestForecastEventType,
			SchemaVersion: event.LatestForecastSchemaVersion,
			Data: event.LatestForecastDataV1{
				LocationID:  int64(index + 1),
				Source:      "open_meteo_best_match",
				RetrievedAt: instant,
			},
		}
		records = append(records, Record{
			Metadata: KafkaRecordMetadata{
				Topic:     event.LatestForecastTopic,
				Partition: 0,
				Offset:    int64(index),
				Key:       forecastEvent.PartitionKey(),
				Timestamp: instant,
			},
			Payload: encodeProcessorEvent(t, forecastEvent),
		})
	}

	if err := processor.ProcessBatch(context.Background(), records); err != nil {
		t.Fatalf("process batch: %v", err)
	}
	if writer.operationalCalls != 1 {
		t.Fatalf("operational writer calls = %d, want 1", writer.operationalCalls)
	}
	if len(writer.operational) != 2 {
		t.Fatalf("operational rows = %d, want 2", len(writer.operational))
	}
	if writer.modelRunCalls != 0 {
		t.Fatalf("model-run writer calls = %d, want 0", writer.modelRunCalls)
	}
}

func TestProcessorMapsEntireBatchBeforeWriting(t *testing.T) {
	instant := testInstant()
	writer := &recordingWriter{}
	processor := newTestProcessor(t, writer, instant)
	forecastEvent := event.LatestForecastEventV1{
		EventID:       "operational-1",
		EventType:     event.LatestForecastEventType,
		SchemaVersion: event.LatestForecastSchemaVersion,
		Data: event.LatestForecastDataV1{
			LocationID:  1,
			Source:      "open_meteo_best_match",
			RetrievedAt: instant,
		},
	}
	records := []Record{
		{
			Metadata: KafkaRecordMetadata{
				Topic:     event.LatestForecastTopic,
				Partition: 0,
				Offset:    0,
				Key:       forecastEvent.PartitionKey(),
				Timestamp: instant,
			},
			Payload: encodeProcessorEvent(t, forecastEvent),
		},
		{
			Metadata: KafkaRecordMetadata{Topic: event.LatestForecastTopic},
			Payload:  []byte("not-json"),
		},
	}

	if err := processor.ProcessBatch(context.Background(), records); err == nil {
		t.Fatal("expected malformed event error")
	}
	if writer.operationalCalls != 0 || writer.modelRunCalls != 0 {
		t.Fatal("partially mapped batch was written")
	}
}

func TestProcessorRejectsUnsupportedTopic(t *testing.T) {
	processor := newTestProcessor(t, &recordingWriter{}, testInstant())

	err := processor.Process(
		context.Background(),
		KafkaRecordMetadata{Topic: "forecast.unknown"},
		[]byte(`{}`),
	)
	if !errors.Is(err, ErrUnsupportedTopic) {
		t.Fatalf("error = %v, want unsupported topic", err)
	}
}

func TestProcessorDoesNotWriteMalformedEvent(t *testing.T) {
	writer := &recordingWriter{}
	processor := newTestProcessor(t, writer, testInstant())

	err := processor.Process(
		context.Background(),
		KafkaRecordMetadata{Topic: event.LatestForecastTopic},
		[]byte("not-json"),
	)
	if err == nil {
		t.Fatal("expected malformed event error")
	}
	if len(writer.operational) != 0 || len(writer.modelRuns) != 0 {
		t.Fatal("malformed event was written")
	}
}

func TestProcessorReturnsWriterError(t *testing.T) {
	instant := testInstant()
	writeError := errors.New("warehouse unavailable")
	writer := &recordingWriter{err: writeError}
	processor := newTestProcessor(t, writer, instant)
	forecastEvent := event.LatestForecastEventV1{
		EventID:       "operational-123",
		EventType:     event.LatestForecastEventType,
		SchemaVersion: event.LatestForecastSchemaVersion,
		Data: event.LatestForecastDataV1{
			LocationID:  3,
			Source:      "open_meteo_best_match",
			RetrievedAt: instant,
		},
	}

	err := processor.Process(
		context.Background(),
		KafkaRecordMetadata{
			Topic:     event.LatestForecastTopic,
			Partition: 0,
			Offset:    0,
			Key:       forecastEvent.PartitionKey(),
			Timestamp: instant,
		},
		encodeProcessorEvent(t, forecastEvent),
	)
	if !errors.Is(err, writeError) {
		t.Fatalf("error = %v, want writer error", err)
	}
}

type recordingWriter struct {
	operational      []OperationalForecastRow
	modelRuns        []ModelRunRow
	operationalCalls int
	modelRunCalls    int
	err              error
}

func (writer *recordingWriter) AppendOperationalForecasts(
	_ context.Context,
	rows []OperationalForecastRow,
) error {
	if writer.err != nil {
		return writer.err
	}
	writer.operationalCalls++
	writer.operational = append(writer.operational, rows...)
	return nil
}

func (writer *recordingWriter) AppendModelRuns(
	_ context.Context,
	rows []ModelRunRow,
) error {
	if writer.err != nil {
		return writer.err
	}
	writer.modelRunCalls++
	writer.modelRuns = append(writer.modelRuns, rows...)
	return nil
}

func newTestProcessor(
	t *testing.T,
	writer Writer,
	now time.Time,
) *Processor {
	t.Helper()

	processor, err := NewProcessor(writer)
	if err != nil {
		t.Fatalf("create processor: %v", err)
	}
	processor.now = func() time.Time { return now }
	return processor
}

func encodeProcessorEvent(t *testing.T, value any) []byte {
	t.Helper()

	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}
	return payload
}
