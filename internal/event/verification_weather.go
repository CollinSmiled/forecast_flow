package event

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

const (
	VerificationWeatherTopic         = "weather.verification"
	VerificationWeatherEventType     = "weather.verification.snapshot"
	VerificationWeatherSchemaVersion = 1
)

type VerificationWeatherEventV1 struct {
	EventID       string                    `json:"event_id"`
	EventType     string                    `json:"event_type"`
	SchemaVersion int                       `json:"schema_version"`
	OccurredAt    time.Time                 `json:"occurred_at"`
	Data          VerificationWeatherDataV1 `json:"data"`
}

type VerificationWeatherDataV1 struct {
	LocationID    int64                         `json:"location_id"`
	Source        string                        `json:"source"`
	ReferenceKind string                        `json:"reference_kind"`
	RetrievedAt   time.Time                     `json:"retrieved_at"`
	Timezone      string                        `json:"timezone"`
	PeriodStart   time.Time                     `json:"period_start"`
	PeriodEnd     time.Time                     `json:"period_end"`
	Hourly        []VerificationHourlyWeatherV1 `json:"hourly"`
}

type VerificationHourlyWeatherV1 struct {
	ValidAt time.Time `json:"valid_at"`
	WeatherMetricsV1
}

func NewVerificationWeatherEventV1(
	snapshot verification.Snapshot,
) (VerificationWeatherEventV1, error) {
	if err := snapshot.Validate(); err != nil {
		return VerificationWeatherEventV1{}, fmt.Errorf(
			"validate verification snapshot: %w",
			err,
		)
	}

	retrievedAt := snapshot.RetrievedAt.UTC()
	hourly := make([]VerificationHourlyWeatherV1, 0, len(snapshot.Hourly))
	for _, item := range snapshot.Hourly {
		hourly = append(hourly, VerificationHourlyWeatherV1{
			ValidAt:          item.ValidAt.UTC(),
			WeatherMetricsV1: mapWeatherMetricsV1(item.WeatherMetrics),
		})
	}

	return VerificationWeatherEventV1{
		EventID:       verificationWeatherEventID(snapshot),
		EventType:     VerificationWeatherEventType,
		SchemaVersion: VerificationWeatherSchemaVersion,
		OccurredAt:    retrievedAt,
		Data: VerificationWeatherDataV1{
			LocationID:    snapshot.LocationID,
			Source:        strings.TrimSpace(snapshot.Source),
			ReferenceKind: snapshot.ReferenceKind,
			RetrievedAt:   retrievedAt,
			Timezone:      strings.TrimSpace(snapshot.Timezone),
			PeriodStart:   snapshot.Hourly[0].ValidAt.UTC(),
			PeriodEnd:     snapshot.Hourly[len(snapshot.Hourly)-1].ValidAt.UTC(),
			Hourly:        hourly,
		},
	}, nil
}

func (weatherEvent VerificationWeatherEventV1) PartitionKey() string {
	return strconv.FormatInt(weatherEvent.Data.LocationID, 10)
}

func verificationWeatherEventID(snapshot verification.Snapshot) string {
	identity := fmt.Sprintf(
		"%s|%s|%d|%s|%s|%s",
		strings.TrimSpace(snapshot.Source),
		snapshot.ReferenceKind,
		snapshot.LocationID,
		snapshot.RetrievedAt.UTC().Format(time.RFC3339Nano),
		snapshot.Hourly[0].ValidAt.UTC().Format(time.RFC3339Nano),
		snapshot.Hourly[len(snapshot.Hourly)-1].ValidAt.UTC().Format(time.RFC3339Nano),
	)
	sum := sha256.Sum256([]byte(identity))

	return hex.EncodeToString(sum[:])
}
