package coldstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
)

type KafkaRecordMetadata struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       string
	Timestamp time.Time
}

type Record struct {
	Metadata KafkaRecordMetadata
	Payload  []byte
}

type OperationalForecastRow struct {
	EventID        string    `bigquery:"event_id" json:"event_id"`
	EventType      string    `bigquery:"event_type" json:"event_type"`
	SchemaVersion  int       `bigquery:"schema_version" json:"schema_version"`
	LocationID     int64     `bigquery:"location_id" json:"location_id"`
	Source         string    `bigquery:"source" json:"source"`
	RetrievedAt    time.Time `bigquery:"retrieved_at" json:"retrieved_at"`
	KafkaTopic     string    `bigquery:"kafka_topic" json:"kafka_topic"`
	KafkaPartition int64     `bigquery:"kafka_partition" json:"kafka_partition"`
	KafkaOffset    int64     `bigquery:"kafka_offset" json:"kafka_offset"`
	KafkaKey       string    `bigquery:"kafka_key" json:"kafka_key"`
	KafkaTimestamp time.Time `bigquery:"kafka_timestamp" json:"kafka_timestamp"`
	Payload        string    `bigquery:"payload" json:"payload"`
	IngestedAt     time.Time `bigquery:"ingested_at" json:"ingested_at"`
}

type ModelRunRow struct {
	EventID        string    `bigquery:"event_id" json:"event_id"`
	EventType      string    `bigquery:"event_type" json:"event_type"`
	SchemaVersion  int       `bigquery:"schema_version" json:"schema_version"`
	LocationID     int64     `bigquery:"location_id" json:"location_id"`
	ModelID        string    `bigquery:"model_id" json:"model_id"`
	Provider       string    `bigquery:"provider" json:"provider"`
	ForecastRunAt  time.Time `bigquery:"forecast_run_at" json:"forecast_run_at"`
	RetrievedAt    time.Time `bigquery:"retrieved_at" json:"retrieved_at"`
	KafkaTopic     string    `bigquery:"kafka_topic" json:"kafka_topic"`
	KafkaPartition int64     `bigquery:"kafka_partition" json:"kafka_partition"`
	KafkaOffset    int64     `bigquery:"kafka_offset" json:"kafka_offset"`
	KafkaKey       string    `bigquery:"kafka_key" json:"kafka_key"`
	KafkaTimestamp time.Time `bigquery:"kafka_timestamp" json:"kafka_timestamp"`
	Payload        string    `bigquery:"payload" json:"payload"`
	IngestedAt     time.Time `bigquery:"ingested_at" json:"ingested_at"`
}

type VerificationWeatherRow struct {
	EventID        string    `bigquery:"event_id" json:"event_id"`
	EventType      string    `bigquery:"event_type" json:"event_type"`
	SchemaVersion  int       `bigquery:"schema_version" json:"schema_version"`
	LocationID     int64     `bigquery:"location_id" json:"location_id"`
	Source         string    `bigquery:"source" json:"source"`
	ReferenceKind  string    `bigquery:"reference_kind" json:"reference_kind"`
	PeriodStart    time.Time `bigquery:"period_start" json:"period_start"`
	PeriodEnd      time.Time `bigquery:"period_end" json:"period_end"`
	RetrievedAt    time.Time `bigquery:"retrieved_at" json:"retrieved_at"`
	KafkaTopic     string    `bigquery:"kafka_topic" json:"kafka_topic"`
	KafkaPartition int64     `bigquery:"kafka_partition" json:"kafka_partition"`
	KafkaOffset    int64     `bigquery:"kafka_offset" json:"kafka_offset"`
	KafkaKey       string    `bigquery:"kafka_key" json:"kafka_key"`
	KafkaTimestamp time.Time `bigquery:"kafka_timestamp" json:"kafka_timestamp"`
	Payload        string    `bigquery:"payload" json:"payload"`
	IngestedAt     time.Time `bigquery:"ingested_at" json:"ingested_at"`
}

func NewOperationalForecastRow(
	forecastEvent event.LatestForecastEventV1,
	metadata KafkaRecordMetadata,
	payload []byte,
	ingestedAt time.Time,
) (OperationalForecastRow, error) {
	if forecastEvent.EventType != event.LatestForecastEventType {
		return OperationalForecastRow{}, fmt.Errorf(
			"unsupported operational event type %q",
			forecastEvent.EventType,
		)
	}

	if forecastEvent.SchemaVersion != event.LatestForecastSchemaVersion {
		return OperationalForecastRow{}, fmt.Errorf(
			"unsupported operational schema version %d",
			forecastEvent.SchemaVersion,
		)
	}

	if forecastEvent.Data.LocationID < 1 {
		return OperationalForecastRow{}, errors.New(
			"operational event location ID is required",
		)
	}

	if strings.TrimSpace(forecastEvent.Data.Source) == "" {
		return OperationalForecastRow{}, errors.New(
			"operational event source is required",
		)
	}

	if forecastEvent.Data.RetrievedAt.IsZero() {
		return OperationalForecastRow{}, errors.New(
			"operational event retrieval time is required",
		)
	}

	if err := validateRawEvent(
		forecastEvent.EventID,
		event.LatestForecastTopic,
		forecastEvent.PartitionKey(),
		metadata,
		payload,
		ingestedAt,
	); err != nil {
		return OperationalForecastRow{}, err
	}

	return OperationalForecastRow{
		EventID:        forecastEvent.EventID,
		EventType:      forecastEvent.EventType,
		SchemaVersion:  forecastEvent.SchemaVersion,
		LocationID:     forecastEvent.Data.LocationID,
		Source:         forecastEvent.Data.Source,
		RetrievedAt:    forecastEvent.Data.RetrievedAt.UTC(),
		KafkaTopic:     metadata.Topic,
		KafkaPartition: int64(metadata.Partition),
		KafkaOffset:    metadata.Offset,
		KafkaKey:       metadata.Key,
		KafkaTimestamp: metadata.Timestamp.UTC(),
		Payload:        string(payload),
		IngestedAt:     ingestedAt.UTC(),
	}, nil
}

func NewModelRunRow(
	forecastEvent event.ForecastRunEventV1,
	metadata KafkaRecordMetadata,
	payload []byte,
	ingestedAt time.Time,
) (ModelRunRow, error) {
	if forecastEvent.EventType != event.ForecastRunEventType {
		return ModelRunRow{}, fmt.Errorf(
			"unsupported model-run event type %q",
			forecastEvent.EventType,
		)
	}

	if forecastEvent.SchemaVersion != event.ForecastRunSchemaVersion {
		return ModelRunRow{}, fmt.Errorf(
			"unsupported model-run schema version %d",
			forecastEvent.SchemaVersion,
		)
	}

	if forecastEvent.Data.LocationID < 1 {
		return ModelRunRow{}, errors.New(
			"model-run event location ID is required",
		)
	}

	if strings.TrimSpace(forecastEvent.Data.Model.ModelID) == "" {
		return ModelRunRow{}, errors.New("model-run model ID is required")
	}

	if strings.TrimSpace(forecastEvent.Data.Model.Provider) == "" {
		return ModelRunRow{}, errors.New("model-run provider is required")
	}

	if forecastEvent.Data.ForecastRunAt.IsZero() {
		return ModelRunRow{}, errors.New("forecast run time is required")
	}

	if forecastEvent.Data.RetrievedAt.IsZero() {
		return ModelRunRow{}, errors.New(
			"model-run retrieval time is required",
		)
	}

	if err := validateRawEvent(
		forecastEvent.EventID,
		event.ForecastRunTopic,
		forecastEvent.PartitionKey(),
		metadata,
		payload,
		ingestedAt,
	); err != nil {
		return ModelRunRow{}, err
	}

	return ModelRunRow{
		EventID:        forecastEvent.EventID,
		EventType:      forecastEvent.EventType,
		SchemaVersion:  forecastEvent.SchemaVersion,
		LocationID:     forecastEvent.Data.LocationID,
		ModelID:        forecastEvent.Data.Model.ModelID,
		Provider:       forecastEvent.Data.Model.Provider,
		ForecastRunAt:  forecastEvent.Data.ForecastRunAt.UTC(),
		RetrievedAt:    forecastEvent.Data.RetrievedAt.UTC(),
		KafkaTopic:     metadata.Topic,
		KafkaPartition: int64(metadata.Partition),
		KafkaOffset:    metadata.Offset,
		KafkaKey:       metadata.Key,
		KafkaTimestamp: metadata.Timestamp.UTC(),
		Payload:        string(payload),
		IngestedAt:     ingestedAt.UTC(),
	}, nil
}

func NewVerificationWeatherRow(
	weatherEvent event.VerificationWeatherEventV1,
	metadata KafkaRecordMetadata,
	payload []byte,
	ingestedAt time.Time,
) (VerificationWeatherRow, error) {
	if weatherEvent.EventType != event.VerificationWeatherEventType {
		return VerificationWeatherRow{}, fmt.Errorf(
			"unsupported verification weather event type %q",
			weatherEvent.EventType,
		)
	}
	if weatherEvent.SchemaVersion != event.VerificationWeatherSchemaVersion {
		return VerificationWeatherRow{}, fmt.Errorf(
			"unsupported verification weather schema version %d",
			weatherEvent.SchemaVersion,
		)
	}
	if weatherEvent.Data.LocationID < 1 {
		return VerificationWeatherRow{}, errors.New(
			"verification weather location ID is required",
		)
	}
	if strings.TrimSpace(weatherEvent.Data.Source) == "" {
		return VerificationWeatherRow{}, errors.New(
			"verification weather source is required",
		)
	}
	if strings.TrimSpace(weatherEvent.Data.ReferenceKind) == "" {
		return VerificationWeatherRow{}, errors.New(
			"verification weather reference kind is required",
		)
	}
	if weatherEvent.Data.PeriodStart.IsZero() || weatherEvent.Data.PeriodEnd.IsZero() {
		return VerificationWeatherRow{}, errors.New(
			"verification weather period is required",
		)
	}
	if weatherEvent.Data.PeriodEnd.Before(weatherEvent.Data.PeriodStart) {
		return VerificationWeatherRow{}, errors.New(
			"verification weather period end must not precede its start",
		)
	}
	if weatherEvent.Data.RetrievedAt.IsZero() {
		return VerificationWeatherRow{}, errors.New(
			"verification weather retrieval time is required",
		)
	}

	if err := validateRawEvent(
		weatherEvent.EventID,
		event.VerificationWeatherTopic,
		weatherEvent.PartitionKey(),
		metadata,
		payload,
		ingestedAt,
	); err != nil {
		return VerificationWeatherRow{}, err
	}

	return VerificationWeatherRow{
		EventID:        weatherEvent.EventID,
		EventType:      weatherEvent.EventType,
		SchemaVersion:  weatherEvent.SchemaVersion,
		LocationID:     weatherEvent.Data.LocationID,
		Source:         weatherEvent.Data.Source,
		ReferenceKind:  weatherEvent.Data.ReferenceKind,
		PeriodStart:    weatherEvent.Data.PeriodStart.UTC(),
		PeriodEnd:      weatherEvent.Data.PeriodEnd.UTC(),
		RetrievedAt:    weatherEvent.Data.RetrievedAt.UTC(),
		KafkaTopic:     metadata.Topic,
		KafkaPartition: int64(metadata.Partition),
		KafkaOffset:    metadata.Offset,
		KafkaKey:       metadata.Key,
		KafkaTimestamp: metadata.Timestamp.UTC(),
		Payload:        string(payload),
		IngestedAt:     ingestedAt.UTC(),
	}, nil
}

func validateRawEvent(
	eventID string,
	expectedTopic string,
	expectedKey string,
	metadata KafkaRecordMetadata,
	payload []byte,
	ingestedAt time.Time,
) error {
	if strings.TrimSpace(eventID) == "" {
		return errors.New("event ID is required")
	}

	if metadata.Topic != expectedTopic {
		return fmt.Errorf(
			"Kafka topic %q does not match expected topic %q",
			metadata.Topic,
			expectedTopic,
		)
	}

	if metadata.Partition < 0 || metadata.Offset < 0 {
		return errors.New("Kafka partition and offset must not be negative")
	}

	if metadata.Key != expectedKey {
		return fmt.Errorf(
			"Kafka key %q does not match event key %q",
			metadata.Key,
			expectedKey,
		)
	}

	if metadata.Timestamp.IsZero() {
		return errors.New("Kafka timestamp is required")
	}

	if ingestedAt.IsZero() {
		return errors.New("ingestion time is required")
	}

	if !json.Valid(payload) {
		return errors.New("raw event payload must contain valid JSON")
	}

	var identity struct {
		EventID string `json:"event_id"`
	}
	if err := json.Unmarshal(payload, &identity); err != nil {
		return fmt.Errorf("read raw event identity: %w", err)
	}

	if identity.EventID != eventID {
		return fmt.Errorf(
			"payload event ID %q does not match event ID %q",
			identity.EventID,
			eventID,
		)
	}

	return nil
}
