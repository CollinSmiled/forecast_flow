package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

const coldPathClientID = "forecast-flow-cold-path"

type ColdPathRecordProcessor interface {
	Process(
		ctx context.Context,
		metadata coldstore.KafkaRecordMetadata,
		payload []byte,
	) error
}

type ColdPathConsumer struct {
	client    *kgo.Client
	processor ColdPathRecordProcessor
	logger    *slog.Logger
}

func NewColdPathConsumer(
	ctx context.Context,
	brokers []string,
	groupID string,
	processor ColdPathRecordProcessor,
	logger *slog.Logger,
) (*ColdPathConsumer, error) {
	brokers = normalizeBrokers(brokers)
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, errors.New("Kafka consumer group ID is required")
	}

	if processor == nil {
		return nil, errors.New("cold-path record processor is required")
	}

	if logger == nil {
		return nil, errors.New("consumer logger is required")
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ClientID(coldPathClientID),
		kgo.ConsumerGroup(groupID),
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
		client:    client,
		processor: processor,
		logger:    logger,
	}, nil
}

func (consumer *ColdPathConsumer) Run(ctx context.Context) error {
	for {
		fetches := consumer.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}

		if fetchErrors := fetches.Errors(); len(fetchErrors) > 0 {
			return fmt.Errorf(
				"poll cold-path records: %w",
				fetchErrors[0].Err,
			)
		}

		iterator := fetches.RecordIter()
		for !iterator.Done() {
			record := iterator.Next()
			if err := consumer.processRecord(ctx, record); err != nil {
				return fmt.Errorf(
					"process cold-path record from topic %s at partition %d offset %d: %w",
					record.Topic,
					record.Partition,
					record.Offset,
					err,
				)
			}

			if err := consumer.client.CommitRecords(ctx, record); err != nil {
				return fmt.Errorf("commit cold-path offset: %w", err)
			}

			consumer.logger.Info(
				"cold-path record processed",
				"topic", record.Topic,
				"partition", record.Partition,
				"offset", record.Offset,
				"key", string(record.Key),
			)
		}
	}
}

func (consumer *ColdPathConsumer) Close() {
	consumer.client.Close()
}

func (consumer *ColdPathConsumer) processRecord(
	ctx context.Context,
	record *kgo.Record,
) error {
	if record == nil {
		return errors.New("Kafka record is required")
	}

	metadata := coldstore.KafkaRecordMetadata{
		Topic:     record.Topic,
		Partition: record.Partition,
		Offset:    record.Offset,
		Key:       string(record.Key),
		Timestamp: record.Timestamp,
	}

	if err := consumer.processor.Process(ctx, metadata, record.Value); err != nil {
		return fmt.Errorf("process cold-path event: %w", err)
	}

	return nil
}
