package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

const hotPathClientID = "forecast-flow-hot-path"

type LatestForecastStore interface {
	ReplaceLatest(
		ctx context.Context,
		forecastEvent event.LatestForecastEventV1,
	) (bool, error)
}

type LatestForecastConsumer struct {
	client *kgo.Client
	store  LatestForecastStore
	logger *slog.Logger
}

func NewLatestForecastConsumer(
	ctx context.Context,
	brokers []string,
	groupID string,
	store LatestForecastStore,
	logger *slog.Logger,
) (*LatestForecastConsumer, error) {
	brokers = normalizeBrokers(brokers)
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, errors.New("Kafka consumer group ID is required")
	}

	if store == nil {
		return nil, errors.New("latest forecast store is required")
	}

	if logger == nil {
		return nil, errors.New("consumer logger is required")
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ClientID(hotPathClientID),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(event.LatestForecastTopic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("create latest forecast consumer: %w", err)
	}

	if err := client.Ping(ctx); err != nil {
		client.Close()

		return nil, fmt.Errorf("ping Kafka broker: %w", err)
	}

	return &LatestForecastConsumer{
		client: client,
		store:  store,
		logger: logger,
	}, nil
}

func (consumer *LatestForecastConsumer) Run(ctx context.Context) error {
	for {
		fetches := consumer.client.PollFetches(ctx)

		if ctx.Err() != nil {
			return nil
		}

		if fetchErrors := fetches.Errors(); len(fetchErrors) > 0 {
			return fmt.Errorf(
				"poll latest forecast records: %w",
				fetchErrors[0].Err,
			)
		}

		iterator := fetches.RecordIter()
		for !iterator.Done() {
			record := iterator.Next()

			forecastEvent, replaced, err := consumer.storeRecord(
				ctx,
				record,
			)
			if err != nil {
				return fmt.Errorf(
					"process latest forecast record at partition %d offset %d: %w",
					record.Partition,
					record.Offset,
					err,
				)
			}

			if err := consumer.client.CommitRecords(ctx, record); err != nil {
				return fmt.Errorf(
					"commit latest forecast offset: %w",
					err,
				)
			}

			consumer.logger.Info(
				"latest forecast record processed",
				"event_id", forecastEvent.EventID,
				"location_id", forecastEvent.Data.LocationID,
				"partition", record.Partition,
				"offset", record.Offset,
				"replaced", replaced,
			)
		}
	}
}

func (consumer *LatestForecastConsumer) Close() {
	consumer.client.Close()
}

func (consumer *LatestForecastConsumer) storeRecord(
	ctx context.Context,
	record *kgo.Record,
) (event.LatestForecastEventV1, bool, error) {
	forecastEvent, err := decodeLatestForecastRecord(record)
	if err != nil {
		return event.LatestForecastEventV1{}, false, err
	}

	replaced, err := consumer.store.ReplaceLatest(ctx, forecastEvent)
	if err != nil {
		return event.LatestForecastEventV1{}, false, fmt.Errorf(
			"store latest forecast: %w",
			err,
		)
	}

	return forecastEvent, replaced, nil
}

func decodeLatestForecastRecord(
	record *kgo.Record,
) (event.LatestForecastEventV1, error) {
	if record == nil {
		return event.LatestForecastEventV1{}, errors.New(
			"Kafka record is required",
		)
	}

	if record.Topic != event.LatestForecastTopic {
		return event.LatestForecastEventV1{}, fmt.Errorf(
			"unexpected Kafka topic %q",
			record.Topic,
		)
	}

	var forecastEvent event.LatestForecastEventV1
	if err := json.Unmarshal(record.Value, &forecastEvent); err != nil {
		return event.LatestForecastEventV1{}, fmt.Errorf(
			"decode latest forecast event: %w",
			err,
		)
	}

	expectedKey := forecastEvent.PartitionKey()
	if string(record.Key) != expectedKey {
		return event.LatestForecastEventV1{}, fmt.Errorf(
			"Kafka record key %q does not match event key %q",
			string(record.Key),
			expectedKey,
		)
	}

	return forecastEvent, nil
}
