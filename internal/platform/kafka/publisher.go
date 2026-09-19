package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

const defaultClientID = "forecast-flow-api"

type Publisher struct {
	client *kgo.Client
}

func NewPublisher(
	ctx context.Context,
	brokers []string,
) (*Publisher, error) {
	brokers = normalizeBrokers(brokers)

	if len(brokers) == 0 {
		return nil, errors.New(
			"at least one Kafka broker is required",
		)
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ClientID(defaultClientID),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		return nil, fmt.Errorf("create Kafka client: %w", err)
	}

	if err := client.Ping(ctx); err != nil {
		client.Close()

		return nil, fmt.Errorf("ping Kafka broker: %w", err)
	}

	return &Publisher{
		client: client,
	}, nil
}

func (publisher *Publisher) PublishLatestForecast(
	ctx context.Context,
	forecastEvent event.LatestForecastEventV1,
) error {
	record, err := latestForecastRecord(forecastEvent)
	if err != nil {
		return err
	}

	if err := publisher.client.
		ProduceSync(ctx, record).
		FirstErr(); err != nil {
		return fmt.Errorf(
			"publish latest forecast event: %w",
			err,
		)
	}

	return nil
}

func (publisher *Publisher) PublishForecastRun(
	ctx context.Context,
	forecastEvent event.ForecastRunEventV1,
) error {
	record, err := forecastRunRecord(forecastEvent)
	if err != nil {
		return err
	}

	if err := publisher.client.
		ProduceSync(ctx, record).
		FirstErr(); err != nil {
		return fmt.Errorf(
			"publish forecast run event: %w",
			err,
		)
	}

	return nil
}

func (publisher *Publisher) Close() {
	publisher.client.Close()
}

func latestForecastRecord(
	forecastEvent event.LatestForecastEventV1,
) (*kgo.Record, error) {
	return newEventRecord(
		event.LatestForecastTopic,
		forecastEvent.PartitionKey(),
		forecastEvent.EventID,
		forecastEvent.EventType,
		forecastEvent.SchemaVersion,
		forecastEvent.OccurredAt,
		forecastEvent,
	)
}

func forecastRunRecord(
	forecastEvent event.ForecastRunEventV1,
) (*kgo.Record, error) {
	return newEventRecord(
		event.ForecastRunTopic,
		forecastEvent.PartitionKey(),
		forecastEvent.EventID,
		forecastEvent.EventType,
		forecastEvent.SchemaVersion,
		forecastEvent.OccurredAt,
		forecastEvent,
	)
}

func newEventRecord(
	topic string,
	partitionKey string,
	eventID string,
	eventType string,
	schemaVersion int,
	occurredAt time.Time,
	payload any,
) (*kgo.Record, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Kafka event: %w", err)
	}

	return &kgo.Record{
		Topic:     topic,
		Key:       []byte(partitionKey),
		Value:     body,
		Timestamp: occurredAt,
		Headers: []kgo.RecordHeader{
			{
				Key:   "event_id",
				Value: []byte(eventID),
			},
			{
				Key:   "event_type",
				Value: []byte(eventType),
			},
			{
				Key: "schema_version",
				Value: []byte(
					strconv.Itoa(
						schemaVersion,
					),
				),
			},
		},
	}, nil
}

func normalizeBrokers(brokers []string) []string {
	normalized := make([]string, 0, len(brokers))

	for _, broker := range brokers {
		broker = strings.TrimSpace(broker)

		if broker != "" {
			normalized = append(normalized, broker)
		}
	}

	return normalized
}
