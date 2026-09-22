package kafka

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestColdPathConsumerMapsKafkaRecordMetadata(t *testing.T) {
	instant := time.Date(2026, time.September, 22, 0, 0, 0, 0, time.UTC)
	processor := &recordingColdPathProcessor{}
	consumer := &ColdPathConsumer{processor: processor}
	record := &kgo.Record{
		Topic:     event.ForecastRunTopic,
		Partition: 2,
		Offset:    17,
		Key:       []byte("3:ecmwf_ifs"),
		Value:     []byte(`{"event_id":"run-123"}`),
		Timestamp: instant,
	}

	if err := consumer.processRecord(context.Background(), record); err != nil {
		t.Fatalf("process record: %v", err)
	}

	if processor.calls != 1 {
		t.Fatalf("processor calls = %d, want 1", processor.calls)
	}
	if processor.metadata.Topic != event.ForecastRunTopic ||
		processor.metadata.Partition != 2 ||
		processor.metadata.Offset != 17 ||
		processor.metadata.Key != "3:ecmwf_ifs" ||
		!processor.metadata.Timestamp.Equal(instant) {
		t.Errorf("metadata = %+v, want Kafka record metadata", processor.metadata)
	}
	if string(processor.payload) != string(record.Value) {
		t.Errorf("payload = %q, want %q", processor.payload, record.Value)
	}
}

func TestColdPathConsumerRejectsNilRecord(t *testing.T) {
	processor := &recordingColdPathProcessor{}
	consumer := &ColdPathConsumer{processor: processor}

	if err := consumer.processRecord(context.Background(), nil); err == nil {
		t.Fatal("expected nil record error")
	}
	if processor.calls != 0 {
		t.Errorf("processor calls = %d, want 0", processor.calls)
	}
}

func TestColdPathConsumerReturnsProcessorError(t *testing.T) {
	processorError := errors.New("warehouse unavailable")
	processor := &recordingColdPathProcessor{err: processorError}
	consumer := &ColdPathConsumer{processor: processor}

	err := consumer.processRecord(
		context.Background(),
		&kgo.Record{Topic: event.LatestForecastTopic},
	)
	if !errors.Is(err, processorError) {
		t.Fatalf("error = %v, want processor error", err)
	}
}

func TestNewColdPathConsumerValidatesConfiguration(t *testing.T) {
	processor := &recordingColdPathProcessor{}
	logger := slog.Default()

	tests := []struct {
		name      string
		brokers   []string
		groupID   string
		processor ColdPathRecordProcessor
		logger    *slog.Logger
	}{
		{
			name:      "brokers",
			groupID:   "cold-path",
			processor: processor,
			logger:    logger,
		},
		{
			name:      "group ID",
			brokers:   []string{"localhost:9092"},
			processor: processor,
			logger:    logger,
		},
		{
			name:    "processor",
			brokers: []string{"localhost:9092"},
			groupID: "cold-path",
			logger:  logger,
		},
		{
			name:      "logger",
			brokers:   []string{"localhost:9092"},
			groupID:   "cold-path",
			processor: processor,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			consumer, err := NewColdPathConsumer(
				context.Background(),
				test.brokers,
				test.groupID,
				test.processor,
				test.logger,
			)
			if err == nil {
				consumer.Close()
				t.Fatal("expected configuration error")
			}
		})
	}
}

type recordingColdPathProcessor struct {
	metadata coldstore.KafkaRecordMetadata
	payload  []byte
	err      error
	calls    int
}

func (processor *recordingColdPathProcessor) Process(
	_ context.Context,
	metadata coldstore.KafkaRecordMetadata,
	payload []byte,
) error {
	processor.calls++
	processor.metadata = metadata
	processor.payload = append([]byte(nil), payload...)
	return processor.err
}
