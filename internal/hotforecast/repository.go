package hotforecast

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
	"github.com/jackc/pgx/v5"
)

type transactionBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	database transactionBeginner
}

func NewRepository(database transactionBeginner) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) ReplaceLatest(
	ctx context.Context,
	forecastEvent event.LatestForecastEventV1,
) (bool, error) {
	hourlyRows, dailyRows, err := replacementRows(forecastEvent)
	if err != nil {
		return false, err
	}

	transaction, err := repository.database.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin latest forecast transaction: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	replaced, err := upsertCurrent(
		ctx,
		transaction,
		forecastEvent,
	)
	if err != nil {
		return false, err
	}

	if !replaced {
		if err := transaction.Commit(ctx); err != nil {
			return false, fmt.Errorf(
				"commit unchanged latest forecast: %w",
				err,
			)
		}

		return false, nil
	}

	locationID := forecastEvent.Data.LocationID

	if _, err := transaction.Exec(
		ctx,
		"DELETE FROM public.latest_operational_hourly WHERE location_id = $1",
		locationID,
	); err != nil {
		return false, fmt.Errorf("delete previous hourly forecast: %w", err)
	}

	if _, err := transaction.Exec(
		ctx,
		"DELETE FROM public.latest_operational_daily WHERE location_id = $1",
		locationID,
	); err != nil {
		return false, fmt.Errorf("delete previous daily forecast: %w", err)
	}

	if _, err := transaction.CopyFrom(
		ctx,
		pgx.Identifier{"public", "latest_operational_hourly"},
		hourlyColumns,
		pgx.CopyFromRows(hourlyRows),
	); err != nil {
		return false, fmt.Errorf("copy latest hourly forecast: %w", err)
	}

	if _, err := transaction.CopyFrom(
		ctx,
		pgx.Identifier{"public", "latest_operational_daily"},
		dailyColumns,
		pgx.CopyFromRows(dailyRows),
	); err != nil {
		return false, fmt.Errorf("copy latest daily forecast: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit latest forecast: %w", err)
	}

	return true, nil
}

func upsertCurrent(
	ctx context.Context,
	transaction pgx.Tx,
	forecastEvent event.LatestForecastEventV1,
) (bool, error) {
	const query = `
		INSERT INTO public.latest_operational_forecasts (
			location_id,
			event_id,
			schema_version,
			source,
			retrieved_at,
			timezone,
			current_valid_at,
			current_interval_seconds,
			current_temperature_2m,
			current_apparent_temperature,
			current_relative_humidity_2m,
			current_precipitation,
			current_precipitation_probability,
			current_rain,
			current_showers,
			current_snowfall,
			current_weather_code,
			current_cloud_cover,
			current_pressure_msl,
			current_visibility,
			current_wind_speed_10m,
			current_wind_direction_10m,
			current_wind_gusts_10m,
			current_uv_index,
			current_is_day
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25
		)
		ON CONFLICT (location_id)
		DO UPDATE SET
			event_id = EXCLUDED.event_id,
			schema_version = EXCLUDED.schema_version,
			source = EXCLUDED.source,
			retrieved_at = EXCLUDED.retrieved_at,
			timezone = EXCLUDED.timezone,
			current_valid_at = EXCLUDED.current_valid_at,
			current_interval_seconds = EXCLUDED.current_interval_seconds,
			current_temperature_2m = EXCLUDED.current_temperature_2m,
			current_apparent_temperature = EXCLUDED.current_apparent_temperature,
			current_relative_humidity_2m = EXCLUDED.current_relative_humidity_2m,
			current_precipitation = EXCLUDED.current_precipitation,
			current_precipitation_probability = EXCLUDED.current_precipitation_probability,
			current_rain = EXCLUDED.current_rain,
			current_showers = EXCLUDED.current_showers,
			current_snowfall = EXCLUDED.current_snowfall,
			current_weather_code = EXCLUDED.current_weather_code,
			current_cloud_cover = EXCLUDED.current_cloud_cover,
			current_pressure_msl = EXCLUDED.current_pressure_msl,
			current_visibility = EXCLUDED.current_visibility,
			current_wind_speed_10m = EXCLUDED.current_wind_speed_10m,
			current_wind_direction_10m = EXCLUDED.current_wind_direction_10m,
			current_wind_gusts_10m = EXCLUDED.current_wind_gusts_10m,
			current_uv_index = EXCLUDED.current_uv_index,
			current_is_day = EXCLUDED.current_is_day,
			updated_at = CURRENT_TIMESTAMP
		WHERE EXCLUDED.retrieved_at >
			public.latest_operational_forecasts.retrieved_at
		RETURNING location_id
	`

	current := forecastEvent.Data.Current
	var locationID int64

	err := transaction.QueryRow(
		ctx,
		query,
		forecastEvent.Data.LocationID,
		forecastEvent.EventID,
		forecastEvent.SchemaVersion,
		forecastEvent.Data.Source,
		forecastEvent.Data.RetrievedAt,
		forecastEvent.Data.Timezone,
		current.ValidAt,
		current.IntervalSeconds,
		current.Temperature2M,
		current.ApparentTemperature,
		current.RelativeHumidity2M,
		current.Precipitation,
		current.PrecipitationProbability,
		current.Rain,
		current.Showers,
		current.Snowfall,
		current.WeatherCode,
		current.CloudCover,
		current.PressureMSL,
		current.Visibility,
		current.WindSpeed10M,
		current.WindDirection10M,
		current.WindGusts10M,
		current.UVIndex,
		current.IsDay,
	).Scan(&locationID)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("upsert latest current forecast: %w", err)
	}

	return true, nil
}

func replacementRows(
	forecastEvent event.LatestForecastEventV1,
) ([][]any, [][]any, error) {
	if strings.TrimSpace(forecastEvent.EventID) == "" {
		return nil, nil, errors.New("latest forecast event ID is required")
	}

	if forecastEvent.EventType != event.LatestForecastEventType {
		return nil, nil, fmt.Errorf(
			"unsupported event type %q",
			forecastEvent.EventType,
		)
	}

	if forecastEvent.SchemaVersion != event.LatestForecastSchemaVersion {
		return nil, nil, fmt.Errorf(
			"unsupported latest forecast schema version %d",
			forecastEvent.SchemaVersion,
		)
	}

	if forecastEvent.Data.LocationID < 1 {
		return nil, nil, errors.New("latest forecast location ID is required")
	}

	if forecastEvent.Data.RetrievedAt.IsZero() {
		return nil, nil, errors.New("latest forecast retrieval time is required")
	}

	if forecastEvent.Data.Current.ValidAt.IsZero() {
		return nil, nil, errors.New("current forecast valid time is required")
	}

	if forecastEvent.Data.Current.IntervalSeconds < 1 {
		return nil, nil, errors.New("current forecast interval must be positive")
	}

	if len(forecastEvent.Data.Hourly) == 0 {
		return nil, nil, errors.New("hourly forecast rows are required")
	}

	if len(forecastEvent.Data.Daily) == 0 {
		return nil, nil, errors.New("daily forecast rows are required")
	}

	hourlyRows := make([][]any, 0, len(forecastEvent.Data.Hourly))
	for index, hourly := range forecastEvent.Data.Hourly {
		if hourly.ValidAt.IsZero() {
			return nil, nil, fmt.Errorf(
				"hourly forecast row %d has no valid time",
				index,
			)
		}

		hourlyRows = append(hourlyRows, hourlyCopyRow(
			forecastEvent.Data.LocationID,
			hourly,
		))
	}

	dailyRows := make([][]any, 0, len(forecastEvent.Data.Daily))
	for index, daily := range forecastEvent.Data.Daily {
		forecastDate, err := time.Parse(time.DateOnly, daily.Date)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"parse daily forecast row %d date: %w",
				index,
				err,
			)
		}

		dailyRows = append(dailyRows, dailyCopyRow(
			forecastEvent.Data.LocationID,
			forecastDate,
			daily,
		))
	}

	return hourlyRows, dailyRows, nil
}

func hourlyCopyRow(
	locationID int64,
	hourly event.OperationalHourlyV1,
) []any {
	return []any{
		locationID,
		hourly.ValidAt,
		hourly.Temperature2M,
		hourly.ApparentTemperature,
		hourly.RelativeHumidity2M,
		hourly.Precipitation,
		hourly.PrecipitationProbability,
		hourly.Rain,
		hourly.Showers,
		hourly.Snowfall,
		hourly.WeatherCode,
		hourly.CloudCover,
		hourly.PressureMSL,
		hourly.Visibility,
		hourly.WindSpeed10M,
		hourly.WindDirection10M,
		hourly.WindGusts10M,
		hourly.UVIndex,
		hourly.IsDay,
	}
}

func dailyCopyRow(
	locationID int64,
	forecastDate time.Time,
	daily event.DailyForecastV1,
) []any {
	return []any{
		locationID,
		forecastDate,
		daily.WeatherCode,
		daily.Temperature2MMax,
		daily.Temperature2MMin,
		daily.ApparentTemperatureMax,
		daily.ApparentTemperatureMin,
		daily.PrecipitationSum,
		daily.PrecipitationProbabilityMax,
		daily.PrecipitationHours,
		daily.WindSpeed10MMax,
		daily.WindGusts10MMax,
		daily.WindDirection10MDominant,
		daily.Sunrise,
		daily.Sunset,
		daily.DaylightDurationSeconds,
		daily.UVIndexMax,
	}
}

var hourlyColumns = []string{
	"location_id",
	"valid_at",
	"temperature_2m",
	"apparent_temperature",
	"relative_humidity_2m",
	"precipitation",
	"precipitation_probability",
	"rain",
	"showers",
	"snowfall",
	"weather_code",
	"cloud_cover",
	"pressure_msl",
	"visibility",
	"wind_speed_10m",
	"wind_direction_10m",
	"wind_gusts_10m",
	"uv_index",
	"is_day",
}

var dailyColumns = []string{
	"location_id",
	"forecast_date",
	"weather_code",
	"temperature_2m_max",
	"temperature_2m_min",
	"apparent_temperature_max",
	"apparent_temperature_min",
	"precipitation_sum",
	"precipitation_probability_max",
	"precipitation_hours",
	"wind_speed_10m_max",
	"wind_gusts_10m_max",
	"wind_direction_10m_dominant",
	"sunrise",
	"sunset",
	"daylight_duration_seconds",
	"uv_index_max",
}
