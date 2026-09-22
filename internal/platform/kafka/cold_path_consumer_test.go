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

func TestColdPathConsumerMapsAndCommitsBatch(t *testing.T) {
	instant := time.Date(2026, time.September, 22, 0, 0, 0, 0, time.UTC)
	processor := &recordingColdPathProcessor{}
	commits := 0
	consumer := &ColdPathConsumer{
		processor: processor,
		client: &fakeColdPathKafkaClient{commit: func(
			_ context.Context,
			records ...*kgo.Record,
		) error {
			commits++
			if len(records) != 2 {
				t.Fatalf("committed records = %d, want 2", len(records))
			}
			return nil
		}},
	}
	records := []*kgo.Record{
		{
			Topic:     event.ForecastRunTopic,
			Partition: 2,
			Offset:    17,
			Key:       []byte("3:ecmwf_ifs"),
			Value:     []byte(`{"event_id":"run-123"}`),
			Timestamp: instant,
		},
		{
			Topic:     event.ForecastRunTopic,
			Partition: 2,
			Offset:    18,
			Key:       []byte("4:ecmwf_ifs"),
			Value:     []byte(`{"event_id":"run-124"}`),
			Timestamp: instant.Add(time.Minute),
		},
	}

	if err := consumer.processBatch(context.Background(), records); err != nil {
		t.Fatalf("process batch: %v", err)
	}

	if processor.calls != 1 {
		t.Fatalf("processor calls = %d, want 1", processor.calls)
	}
	if len(processor.records) != 2 {
		t.Fatalf("processor records = %d, want 2", len(processor.records))
	}
	metadata := processor.records[0].Metadata
	if metadata.Topic != event.ForecastRunTopic ||
		metadata.Partition != 2 ||
		metadata.Offset != 17 ||
		metadata.Key != "3:ecmwf_ifs" ||
		!metadata.Timestamp.Equal(instant) {
		t.Errorf("metadata = %+v, want Kafka record metadata", metadata)
	}
	if string(processor.records[0].Payload) != string(records[0].Value) {
		t.Errorf(
			"payload = %q, want %q",
			processor.records[0].Payload,
			records[0].Value,
		)
	}
	if commits != 1 {
		t.Fatalf("commit calls = %d, want 1", commits)
	}
}

func TestColdPathConsumerRejectsNilRecordWithoutCommit(t *testing.T) {
	processor := &recordingColdPathProcessor{}
	commits := 0
	consumer := &ColdPathConsumer{
		processor: processor,
		client: &fakeColdPathKafkaClient{commit: func(
			context.Context,
			...*kgo.Record,
		) error {
			commits++
			return nil
		}},
	}

	if err := consumer.processBatch(
		context.Background(),
		[]*kgo.Record{nil},
	); err == nil {
		t.Fatal("expected nil record error")
	}
	if processor.calls != 0 {
		t.Errorf("processor calls = %d, want 0", processor.calls)
	}
	if commits != 0 {
		t.Errorf("commit calls = %d, want 0", commits)
	}
}

func TestColdPathConsumerDoesNotCommitProcessorFailure(t *testing.T) {
	processorError := errors.New("warehouse unavailable")
	processor := &recordingColdPathProcessor{err: processorError}
	commits := 0
	consumer := &ColdPathConsumer{
		processor: processor,
		client: &fakeColdPathKafkaClient{commit: func(
			context.Context,
			...*kgo.Record,
		) error {
			commits++
			return nil
		}},
	}

	err := consumer.processBatch(
		context.Background(),
		[]*kgo.Record{{Topic: event.LatestForecastTopic}},
	)
	if !errors.Is(err, processorError) {
		t.Fatalf("error = %v, want processor error", err)
	}
	if commits != 0 {
		t.Fatalf("commit calls = %d, want 0", commits)
	}
}

func TestColdPathConsumerProcessesAndCommitsTopicsIndependently(t *testing.T) {
	processor := &recordingColdPathProcessor{}
	committedTopics := make([]string, 0, 2)
	consumer := &ColdPathConsumer{
		processor: processor,
		client: &fakeColdPathKafkaClient{commit: func(
			_ context.Context,
			records ...*kgo.Record,
		) error {
			committedTopics = append(committedTopics, records[0].Topic)
			return nil
		}},
	}

	err := consumer.processBatch(context.Background(), []*kgo.Record{
		{Topic: event.LatestForecastTopic},
		{Topic: event.ForecastRunTopic},
	})
	if err != nil {
		t.Fatalf("process batch: %v", err)
	}
	if processor.calls != 2 {
		t.Fatalf("processor calls = %d, want 2", processor.calls)
	}
	if len(committedTopics) != 2 ||
		committedTopics[0] != event.LatestForecastTopic ||
		committedTopics[1] != event.ForecastRunTopic {
		t.Fatalf("committed topics = %v, want both topics in fetch order", committedTopics)
	}
}

func TestColdPathConsumerReturnsCommitFailure(t *testing.T) {
	commitError := errors.New("commit failed")
	consumer := &ColdPathConsumer{
		processor: &recordingColdPathProcessor{},
		client: &fakeColdPathKafkaClient{commit: func(
			context.Context,
			...*kgo.Record,
		) error {
			return commitError
		}},
	}

	err := consumer.processBatch(
		context.Background(),
		[]*kgo.Record{{Topic: event.LatestForecastTopic}},
	)
	if !errors.Is(err, commitError) {
		t.Fatalf("error = %v, want commit error", err)
	}
}

func TestColdPathConsumerCollectsUpToMaximumBatchSize(t *testing.T) {
	client := &fakeColdPathKafkaClient{
		poll: func(_ context.Context, maxRecords int) kgo.Fetches {
			records := []*kgo.Record{
				{Topic: event.LatestForecastTopic, Offset: 1},
				{Topic: event.LatestForecastTopic, Offset: 2},
			}
			if maxRecords < len(records) {
				records = records[:maxRecords]
			}
			return fetchesWithRecords(event.LatestForecastTopic, records)
		},
	}
	consumer := &ColdPathConsumer{
		client:          client,
		batchInterval:   time.Hour,
		maxBatchRecords: 2,
		pollTimeout:     time.Second,
	}

	records, err := consumer.collectBatch(context.Background())
	if err != nil {
		t.Fatalf("collect batch: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
	if client.pollCalls != 1 {
		t.Fatalf("poll calls = %d, want 1", client.pollCalls)
	}
}

func TestNewColdPathConsumerValidatesConfiguration(t *testing.T) {
	validConfig := ColdPathConsumerConfig{
		Brokers:         []string{"localhost:9092"},
		GroupID:         "cold-path",
		BatchInterval:   time.Hour,
		MaxBatchRecords: 500,
	}
	processor := &recordingColdPathProcessor{}
	logger := slog.Default()

	tests := []struct {
		name      string
		config    ColdPathConsumerConfig
		processor ColdPathRecordProcessor
		logger    *slog.Logger
	}{
		{
			name: "brokers",
			config: ColdPathConsumerConfig{
				GroupID:         validConfig.GroupID,
				BatchInterval:   validConfig.BatchInterval,
				MaxBatchRecords: validConfig.MaxBatchRecords,
			},
			processor: processor,
			logger:    logger,
		},
		{
			name: "group ID",
			config: ColdPathConsumerConfig{
				Brokers:         validConfig.Brokers,
				BatchInterval:   validConfig.BatchInterval,
				MaxBatchRecords: validConfig.MaxBatchRecords,
			},
			processor: processor,
			logger:    logger,
		},
		{
			name: "batch interval",
			config: ColdPathConsumerConfig{
				Brokers:         validConfig.Brokers,
				GroupID:         validConfig.GroupID,
				MaxBatchRecords: validConfig.MaxBatchRecords,
			},
			processor: processor,
			logger:    logger,
		},
		{
			name: "batch records",
			config: ColdPathConsumerConfig{
				Brokers:       validConfig.Brokers,
				GroupID:       validConfig.GroupID,
				BatchInterval: validConfig.BatchInterval,
			},
			processor: processor,
			logger:    logger,
		},
		{
			name:   "processor",
			config: validConfig,
			logger: logger,
		},
		{
			name:      "logger",
			config:    validConfig,
			processor: processor,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			consumer, err := NewColdPathConsumer(
				context.Background(),
				test.config,
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
	records []coldstore.Record
	err     error
	calls   int
}

type fakeColdPathKafkaClient struct {
	poll      func(context.Context, int) kgo.Fetches
	commit    func(context.Context, ...*kgo.Record) error
	pollCalls int
}

func (client *fakeColdPathKafkaClient) PollRecords(
	ctx context.Context,
	maxRecords int,
) kgo.Fetches {
	client.pollCalls++
	if client.poll == nil {
		return nil
	}
	return client.poll(ctx, maxRecords)
}

func (client *fakeColdPathKafkaClient) CommitRecords(
	ctx context.Context,
	records ...*kgo.Record,
) error {
	if client.commit == nil {
		return nil
	}
	return client.commit(ctx, records...)
}

func (client *fakeColdPathKafkaClient) Close() {}

func fetchesWithRecords(topic string, records []*kgo.Record) kgo.Fetches {
	return kgo.Fetches{{
		Topics: []kgo.FetchTopic{{
			Topic: topic,
			Partitions: []kgo.FetchPartition{{
				Records: records,
			}},
		}},
	}}
}

func (processor *recordingColdPathProcessor) ProcessBatch(
	_ context.Context,
	records []coldstore.Record,
) error {
	processor.calls++
	processor.records = append([]coldstore.Record(nil), records...)
	return processor.err
}
