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
	LatestForecastTopic         = "forecast.latest"
	LatestForecastEventType     = "forecast.latest.snapshot"
	LatestForecastSchemaVersion = 1
)

type LatestForecastEventV1 struct {
	EventID       string               `json:"event_id"`
	EventType     string               `json:"event_type"`
	SchemaVersion int                  `json:"schema_version"`
	OccurredAt    time.Time            `json:"occurred_at"`
	Data          LatestForecastDataV1 `json:"data"`
}

type LatestForecastDataV1 struct {
	LocationID  int64                 `json:"location_id"`
	Source      string                `json:"source"`
	RetrievedAt time.Time             `json:"retrieved_at"`
	Timezone    string                `json:"timezone"`
	Current     CurrentConditionsV1   `json:"current"`
	Hourly      []OperationalHourlyV1 `json:"hourly"`
	Daily       []DailyForecastV1     `json:"daily"`
}

type CurrentConditionsV1 struct {
	ValidAt         time.Time `json:"valid_at"`
	IntervalSeconds int       `json:"interval_seconds"`
	WeatherMetricsV1
}

type OperationalHourlyV1 struct {
	ValidAt time.Time `json:"valid_at"`
	WeatherMetricsV1
}

type WeatherMetricsV1 struct {
	Temperature2M            *float64 `json:"temperature_2m"`
	ApparentTemperature      *float64 `json:"apparent_temperature"`
	RelativeHumidity2M       *float64 `json:"relative_humidity_2m"`
	Precipitation            *float64 `json:"precipitation"`
	PrecipitationProbability *float64 `json:"precipitation_probability"`
	Rain                     *float64 `json:"rain"`
	Showers                  *float64 `json:"showers"`
	Snowfall                 *float64 `json:"snowfall"`
	WeatherCode              *int     `json:"weather_code"`
	CloudCover               *float64 `json:"cloud_cover"`
	PressureMSL              *float64 `json:"pressure_msl"`
	Visibility               *float64 `json:"visibility"`
	WindSpeed10M             *float64 `json:"wind_speed_10m"`
	WindDirection10M         *float64 `json:"wind_direction_10m"`
	WindGusts10M             *float64 `json:"wind_gusts_10m"`
	UVIndex                  *float64 `json:"uv_index"`
	IsDay                    *bool    `json:"is_day"`
}

type DailyForecastV1 struct {
	Date                        string     `json:"date"`
	WeatherCode                 *int       `json:"weather_code"`
	Temperature2MMax            *float64   `json:"temperature_2m_max"`
	Temperature2MMin            *float64   `json:"temperature_2m_min"`
	ApparentTemperatureMax      *float64   `json:"apparent_temperature_max"`
	ApparentTemperatureMin      *float64   `json:"apparent_temperature_min"`
	PrecipitationSum            *float64   `json:"precipitation_sum"`
	PrecipitationProbabilityMax *float64   `json:"precipitation_probability_max"`
	PrecipitationHours          *float64   `json:"precipitation_hours"`
	WindSpeed10MMax             *float64   `json:"wind_speed_10m_max"`
	WindGusts10MMax             *float64   `json:"wind_gusts_10m_max"`
	WindDirection10MDominant    *float64   `json:"wind_direction_10m_dominant"`
	Sunrise                     *time.Time `json:"sunrise"`
	Sunset                      *time.Time `json:"sunset"`
	DaylightDurationSeconds     *float64   `json:"daylight_duration_seconds"`
	UVIndexMax                  *float64   `json:"uv_index_max"`
}

func NewLatestForecastEventV1(
	snapshot forecast.OperationalForecastSnapshot,
) (LatestForecastEventV1, error) {
	if snapshot.LocationID <= 0 {
		return LatestForecastEventV1{}, fmt.Errorf(
			"location ID must be greater than zero",
		)
	}

	if strings.TrimSpace(snapshot.Source) == "" {
		return LatestForecastEventV1{}, fmt.Errorf(
			"forecast source is required",
		)
	}

	if snapshot.RetrievedAt.IsZero() {
		return LatestForecastEventV1{}, fmt.Errorf(
			"retrieval time is required",
		)
	}

	if strings.TrimSpace(snapshot.Timezone) == "" {
		return LatestForecastEventV1{}, fmt.Errorf(
			"timezone is required",
		)
	}

	if snapshot.Current.ValidAt.IsZero() {
		return LatestForecastEventV1{}, fmt.Errorf(
			"current valid time is required",
		)
	}

	if len(snapshot.Hourly) == 0 {
		return LatestForecastEventV1{}, fmt.Errorf(
			"hourly forecasts are required",
		)
	}

	if len(snapshot.Daily) == 0 {
		return LatestForecastEventV1{}, fmt.Errorf(
			"daily forecasts are required",
		)
	}

	retrievedAt := snapshot.RetrievedAt.UTC()

	return LatestForecastEventV1{
		EventID:       latestForecastEventID(snapshot),
		EventType:     LatestForecastEventType,
		SchemaVersion: LatestForecastSchemaVersion,
		OccurredAt:    retrievedAt,
		Data: LatestForecastDataV1{
			LocationID:  snapshot.LocationID,
			Source:      snapshot.Source,
			RetrievedAt: retrievedAt,
			Timezone:    snapshot.Timezone,
			Current:     mapCurrentConditionsV1(snapshot.Current),
			Hourly:      mapOperationalHourlyV1(snapshot.Hourly),
			Daily:       mapDailyForecastV1(snapshot.Daily),
		},
	}, nil
}

func (event LatestForecastEventV1) PartitionKey() string {
	return strconv.FormatInt(event.Data.LocationID, 10)
}

func latestForecastEventID(
	snapshot forecast.OperationalForecastSnapshot,
) string {
	identity := fmt.Sprintf(
		"%s|%d|%s",
		snapshot.Source,
		snapshot.LocationID,
		snapshot.RetrievedAt.UTC().Format(time.RFC3339Nano),
	)
	sum := sha256.Sum256([]byte(identity))

	return hex.EncodeToString(sum[:])
}

func mapCurrentConditionsV1(
	current forecast.CurrentConditions,
) CurrentConditionsV1 {
	return CurrentConditionsV1{
		ValidAt:         current.ValidAt.UTC(),
		IntervalSeconds: current.IntervalSeconds,
		WeatherMetricsV1: mapWeatherMetricsV1(
			current.WeatherMetrics,
		),
	}
}

func mapOperationalHourlyV1(
	hourly []forecast.OperationalHourlyForecast,
) []OperationalHourlyV1 {
	result := make([]OperationalHourlyV1, 0, len(hourly))

	for _, item := range hourly {
		result = append(result, OperationalHourlyV1{
			ValidAt:          item.ValidAt.UTC(),
			WeatherMetricsV1: mapWeatherMetricsV1(item.WeatherMetrics),
		})
	}

	return result
}

func mapWeatherMetricsV1(
	metrics forecast.WeatherMetrics,
) WeatherMetricsV1 {
	return WeatherMetricsV1{
		Temperature2M:            metrics.Temperature2M,
		ApparentTemperature:      metrics.ApparentTemperature,
		RelativeHumidity2M:       metrics.RelativeHumidity2M,
		Precipitation:            metrics.Precipitation,
		PrecipitationProbability: metrics.PrecipitationProbability,
		Rain:                     metrics.Rain,
		Showers:                  metrics.Showers,
		Snowfall:                 metrics.Snowfall,
		WeatherCode:              metrics.WeatherCode,
		CloudCover:               metrics.CloudCover,
		PressureMSL:              metrics.PressureMSL,
		Visibility:               metrics.Visibility,
		WindSpeed10M:             metrics.WindSpeed10M,
		WindDirection10M:         metrics.WindDirection10M,
		WindGusts10M:             metrics.WindGusts10M,
		UVIndex:                  metrics.UVIndex,
		IsDay:                    metrics.IsDay,
	}
}

func mapDailyForecastV1(
	daily []forecast.DailyForecast,
) []DailyForecastV1 {
	result := make([]DailyForecastV1, 0, len(daily))

	for _, item := range daily {
		result = append(result, DailyForecastV1{
			Date:                        item.Date,
			WeatherCode:                 item.WeatherCode,
			Temperature2MMax:            item.Temperature2MMax,
			Temperature2MMin:            item.Temperature2MMin,
			ApparentTemperatureMax:      item.ApparentTemperatureMax,
			ApparentTemperatureMin:      item.ApparentTemperatureMin,
			PrecipitationSum:            item.PrecipitationSum,
			PrecipitationProbabilityMax: item.PrecipitationProbabilityMax,
			PrecipitationHours:          item.PrecipitationHours,
			WindSpeed10MMax:             item.WindSpeed10MMax,
			WindGusts10MMax:             item.WindGusts10MMax,
			WindDirection10MDominant:    item.WindDirection10MDominant,
			Sunrise:                     utcTimePointer(item.Sunrise),
			Sunset:                      utcTimePointer(item.Sunset),
			DaylightDurationSeconds:     item.DaylightDurationSeconds,
			UVIndexMax:                  item.UVIndexMax,
		})
	}

	return result
}

func utcTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	utc := value.UTC()
	return &utc
}
