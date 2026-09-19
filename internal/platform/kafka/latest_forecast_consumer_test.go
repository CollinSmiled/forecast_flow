package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestLatestForecastConsumerStoresDecodedRecord(t *testing.T) {
	forecastEvent := consumerTestEvent()
	record := consumerTestRecord(t, forecastEvent)
	store := &fakeLatestForecastStore{replaced: true}
	consumer := &LatestForecastConsumer{
		store:  store,
		logger: slog.Default(),
	}

	storedEvent, replaced, err := consumer.storeRecord(
		context.Background(),
		record,
	)
	if err != nil {
		t.Fatalf("store record: %v", err)
	}

	if !replaced {
		t.Fatal("record was not reported as replaced")
	}

	if store.calls != 1 {
		t.Errorf("store calls = %d, want 1", store.calls)
	}

	if storedEvent.EventID != forecastEvent.EventID {
		t.Errorf(
			"event ID = %q, want %q",
			storedEvent.EventID,
			forecastEvent.EventID,
		)
	}
}

func TestLatestForecastConsumerRejectsMalformedJSON(t *testing.T) {
	store := &fakeLatestForecastStore{}
	consumer := &LatestForecastConsumer{store: store}

	_, _, err := consumer.storeRecord(
		context.Background(),
		&kgo.Record{
			Topic: event.LatestForecastTopic,
			Key:   []byte("3"),
			Value: []byte("not-json"),
		},
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if store.calls != 0 {
		t.Errorf("store calls = %d, want 0", store.calls)
	}
}

func TestLatestForecastConsumerRejectsMismatchedKey(t *testing.T) {
	forecastEvent := consumerTestEvent()
	record := consumerTestRecord(t, forecastEvent)
	record.Key = []byte("999")

	store := &fakeLatestForecastStore{}
	consumer := &LatestForecastConsumer{store: store}

	_, _, err := consumer.storeRecord(
		context.Background(),
		record,
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if store.calls != 0 {
		t.Errorf("store calls = %d, want 0", store.calls)
	}
}

func TestLatestForecastConsumerReturnsStoreError(t *testing.T) {
	storeError := errors.New("database unavailable")
	store := &fakeLatestForecastStore{err: storeError}
	consumer := &LatestForecastConsumer{store: store}

	_, _, err := consumer.storeRecord(
		context.Background(),
		consumerTestRecord(t, consumerTestEvent()),
	)
	if !errors.Is(err, storeError) {
		t.Fatalf("error = %v, want wrapped store error", err)
	}
}

type fakeLatestForecastStore struct {
	replaced bool
	err      error
	calls    int
}

func (store *fakeLatestForecastStore) ReplaceLatest(
	_ context.Context,
	_ event.LatestForecastEventV1,
) (bool, error) {
	store.calls++

	return store.replaced, store.err
}

func consumerTestEvent() event.LatestForecastEventV1 {
	return event.LatestForecastEventV1{
		EventID:       "event-123",
		EventType:     event.LatestForecastEventType,
		SchemaVersion: event.LatestForecastSchemaVersion,
		OccurredAt: time.Date(
			2026,
			time.September,
			19,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		Data: event.LatestForecastDataV1{
			LocationID: 3,
		},
	}
}

func consumerTestRecord(
	t *testing.T,
	forecastEvent event.LatestForecastEventV1,
) *kgo.Record {
	t.Helper()

	body, err := json.Marshal(forecastEvent)
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}

	return &kgo.Record{
		Topic: event.LatestForecastTopic,
		Key:   []byte(forecastEvent.PartitionKey()),
		Value: body,
	}
}
