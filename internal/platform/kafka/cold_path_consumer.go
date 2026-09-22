package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	coldPathClientID   = "forecast-flow-cold-path"
	defaultPollTimeout = time.Second
)

type ColdPathConsumerConfig struct {
	Brokers         []string
	GroupID         string
	BatchInterval   time.Duration
	MaxBatchRecords int
}

type ColdPathRecordProcessor interface {
	ProcessBatch(ctx context.Context, records []coldstore.Record) error
}

type coldPathKafkaClient interface {
	PollRecords(ctx context.Context, maxRecords int) kgo.Fetches
	CommitRecords(ctx context.Context, records ...*kgo.Record) error
	Close()
}

type ColdPathConsumer struct {
	client          coldPathKafkaClient
	processor       ColdPathRecordProcessor
	logger          *slog.Logger
	batchInterval   time.Duration
	maxBatchRecords int
	pollTimeout     time.Duration
}

func NewColdPathConsumer(
	ctx context.Context,
	config ColdPathConsumerConfig,
	processor ColdPathRecordProcessor,
	logger *slog.Logger,
) (*ColdPathConsumer, error) {
	config.Brokers = normalizeBrokers(config.Brokers)
	if len(config.Brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}

	config.GroupID = strings.TrimSpace(config.GroupID)
	if config.GroupID == "" {
		return nil, errors.New("Kafka consumer group ID is required")
	}

	if config.BatchInterval <= 0 {
		return nil, errors.New("cold-path batch interval must be greater than zero")
	}

	if config.MaxBatchRecords < 1 {
		return nil, errors.New("cold-path maximum batch records must be greater than zero")
	}

	if processor == nil {
		return nil, errors.New("cold-path record processor is required")
	}

	if logger == nil {
		return nil, errors.New("consumer logger is required")
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.Brokers...),
		kgo.ClientID(coldPathClientID),
		kgo.ConsumerGroup(config.GroupID),
		kgo.ConsumeTopics(
			event.LatestForecastTopic,
			event.ForecastRunTopic,
		),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("create cold-path consumer: %w", err)
	}

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping Kafka broker: %w", err)
	}

	return &ColdPathConsumer{
		client:          client,
		processor:       processor,
		logger:          logger,
		batchInterval:   config.BatchInterval,
		maxBatchRecords: config.MaxBatchRecords,
		pollTimeout:     defaultPollTimeout,
	}, nil
}

func (consumer *ColdPathConsumer) Run(ctx context.Context) error {
	for {
		records, err := consumer.collectBatch(ctx)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
		if len(records) == 0 {
			continue
		}

		if err := consumer.processBatch(ctx, records); err != nil {
			return err
		}

		consumer.logger.Info(
			"cold-path batch processed",
			"records", len(records),
		)
	}
}

func (consumer *ColdPathConsumer) Close() {
	consumer.client.Close()
}

func (consumer *ColdPathConsumer) collectBatch(
	ctx context.Context,
) ([]*kgo.Record, error) {
	deadline := time.Now().Add(consumer.batchInterval)
	records := make([]*kgo.Record, 0, consumer.maxBatchRecords)

	for len(records) < consumer.maxBatchRecords {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return records, nil
		}

		pollFor := min(consumer.pollTimeout, remaining)
		pollContext, cancel := context.WithTimeout(ctx, pollFor)
		fetches := consumer.client.PollRecords(
			pollContext,
			consumer.maxBatchRecords-len(records),
		)
		pollError := pollContext.Err()
		cancel()

		if ctx.Err() != nil {
			return nil, nil
		}

		for _, fetchError := range fetches.Errors() {
			if errors.Is(pollError, context.DeadlineExceeded) &&
				errors.Is(fetchError.Err, context.DeadlineExceeded) {
				continue
			}
			return nil, fmt.Errorf(
				"poll cold-path records from topic %s partition %d: %w",
				fetchError.Topic,
				fetchError.Partition,
				fetchError.Err,
			)
		}

		iterator := fetches.RecordIter()
		for !iterator.Done() {
			records = append(records, iterator.Next())
		}
	}

	return records, nil
}

func (consumer *ColdPathConsumer) processBatch(
	ctx context.Context,
	records []*kgo.Record,
) error {
	if len(records) == 0 {
		return nil
	}

	type topicBatch struct {
		kafkaRecords []*kgo.Record
		coldRecords  []coldstore.Record
	}

	topicOrder := make([]string, 0, 2)
	topicBatches := make(map[string]*topicBatch, 2)
	for _, record := range records {
		if record == nil {
			return errors.New("Kafka record is required")
		}

		batch, exists := topicBatches[record.Topic]
		if !exists {
			batch = &topicBatch{}
			topicBatches[record.Topic] = batch
			topicOrder = append(topicOrder, record.Topic)
		}

		batch.kafkaRecords = append(batch.kafkaRecords, record)
		batch.coldRecords = append(batch.coldRecords, coldstore.Record{
			Metadata: coldstore.KafkaRecordMetadata{
				Topic:     record.Topic,
				Partition: record.Partition,
				Offset:    record.Offset,
				Key:       string(record.Key),
				Timestamp: record.Timestamp,
			},
			Payload: record.Value,
		})
	}

	for _, topic := range topicOrder {
		batch := topicBatches[topic]
		if err := consumer.processor.ProcessBatch(ctx, batch.coldRecords); err != nil {
			return fmt.Errorf("process cold-path batch for topic %s: %w", topic, err)
		}

		if err := consumer.client.CommitRecords(ctx, batch.kafkaRecords...); err != nil {
			return fmt.Errorf(
				"commit cold-path batch offsets for topic %s: %w",
				topic,
				err,
			)
		}
	}

	return nil
}
