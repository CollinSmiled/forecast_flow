package event

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
)

const (
	ForecastRunTopic         = "forecast.runs"
	ForecastRunEventType     = "forecast.model_run.snapshot"
	ForecastRunSchemaVersion = 1
)

type ForecastRunEventV1 struct {
	EventID       string            `json:"event_id"`
	EventType     string            `json:"event_type"`
	SchemaVersion int               `json:"schema_version"`
	OccurredAt    time.Time         `json:"occurred_at"`
	Data          ForecastRunDataV1 `json:"data"`
}

type ForecastRunDataV1 struct {
	LocationID    int64                       `json:"location_id"`
	Model         ForecastModelV1             `json:"model"`
	ForecastRunAt time.Time                   `json:"forecast_run_at"`
	RetrievedAt   time.Time                   `json:"retrieved_at"`
	Hourly        []ForecastRunHourlyRecordV1 `json:"hourly"`
}

type ForecastModelV1 struct {
	ModelID      string  `json:"model_id"`
	ModelName    string  `json:"model_name"`
	Provider     string  `json:"provider"`
	ResolutionKM float64 `json:"resolution_km"`
}

type ForecastRunHourlyRecordV1 struct {
	ValidAt       time.Time `json:"valid_at"`
	LeadTimeHours int       `json:"lead_time_hours"`
	WeatherMetricsV1
}

func NewForecastRunEventV1(
	forecastRun forecast.ForecastRun,
) (ForecastRunEventV1, error) {
	run := forecastRun.Run

	if run.LocationID < 1 {
		return ForecastRunEventV1{}, fmt.Errorf(
			"location ID must be greater than zero",
		)
	}

	if err := run.Model.Validate(); err != nil {
		return ForecastRunEventV1{}, fmt.Errorf("validate model: %w", err)
	}

	if run.ForecastRunAt.IsZero() {
		return ForecastRunEventV1{}, fmt.Errorf(
			"forecast run time is required",
		)
	}

	if run.RetrievedAt.IsZero() {
		return ForecastRunEventV1{}, fmt.Errorf(
			"retrieval time is required",
		)
	}

	if len(forecastRun.Hourly) == 0 {
		return ForecastRunEventV1{}, fmt.Errorf(
			"hourly forecasts are required",
		)
	}

	hourly, err := mapForecastRunHourlyV1(run, forecastRun.Hourly)
	if err != nil {
		return ForecastRunEventV1{}, err
	}

	return ForecastRunEventV1{
		EventID:       forecastRunEventID(run),
		EventType:     ForecastRunEventType,
		SchemaVersion: ForecastRunSchemaVersion,
		OccurredAt:    run.RetrievedAt.UTC(),
		Data: ForecastRunDataV1{
			LocationID: run.LocationID,
			Model: ForecastModelV1{
				ModelID:      run.Model.ID,
				ModelName:    run.Model.Name,
				Provider:     run.Model.Provider,
				ResolutionKM: run.Model.ResolutionKM,
			},
			ForecastRunAt: run.ForecastRunAt.UTC(),
			RetrievedAt:   run.RetrievedAt.UTC(),
			Hourly:        hourly,
		},
	}, nil
}

func (forecastEvent ForecastRunEventV1) PartitionKey() string {
	return strconv.FormatInt(forecastEvent.Data.LocationID, 10) +
		":" + strings.TrimSpace(forecastEvent.Data.Model.ModelID)
}

func forecastRunEventID(run forecast.Run) string {
	identity := fmt.Sprintf(
		"%s|%s|%d|%s",
		strings.TrimSpace(run.Model.Provider),
		strings.TrimSpace(run.Model.ID),
		run.LocationID,
		run.ForecastRunAt.UTC().Format(time.RFC3339Nano),
	)
	sum := sha256.Sum256([]byte(identity))

	return hex.EncodeToString(sum[:])
}

func mapForecastRunHourlyV1(
	run forecast.Run,
	hourly []forecast.HourlyForecast,
) ([]ForecastRunHourlyRecordV1, error) {
	result := make([]ForecastRunHourlyRecordV1, 0, len(hourly))

	for index, item := range hourly {
		expectedLeadTime, err := run.LeadTimeHours(item.ValidAt)
		if err != nil {
			return nil, fmt.Errorf(
				"validate hourly forecast at index %d: %w",
				index,
				err,
			)
		}

		if item.LeadTimeHours != expectedLeadTime {
			return nil, fmt.Errorf(
				"hourly forecast at index %d has lead time %d; expected %d",
				index,
				item.LeadTimeHours,
				expectedLeadTime,
			)
		}

		result = append(result, ForecastRunHourlyRecordV1{
			ValidAt:          item.ValidAt.UTC(),
			LeadTimeHours:    item.LeadTimeHours,
			WeatherMetricsV1: mapWeatherMetricsV1(item.WeatherMetrics),
		})
	}

	return result, nil
}
