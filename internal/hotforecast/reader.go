package hotforecast

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/jackc/pgx/v5"
)

var ErrLatestForecastNotFound = errors.New("latest forecast not found")

type Latest struct {
	EventID       string
	SchemaVersion int
	Location      location.Location
	Snapshot      forecast.OperationalForecastSnapshot
}

func (repository *Repository) GetLatest(
	ctx context.Context,
	locationID int64,
) (Latest, error) {
	if locationID < 1 {
		return Latest{}, errors.New(
			"get latest forecast: location ID must be greater than zero",
		)
	}

	transaction, err := repository.database.Begin(ctx)
	if err != nil {
		return Latest{}, fmt.Errorf("begin latest forecast read: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	latest, err := readCurrent(ctx, transaction, locationID)
	if err != nil {
		return Latest{}, err
	}

	latest.Snapshot.Hourly, err = readHourly(
		ctx,
		transaction,
		locationID,
	)
	if err != nil {
		return Latest{}, err
	}

	latest.Snapshot.Daily, err = readDaily(
		ctx,
		transaction,
		locationID,
	)
	if err != nil {
		return Latest{}, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return Latest{}, fmt.Errorf("commit latest forecast read: %w", err)
	}

	return latest, nil
}

func readCurrent(
	ctx context.Context,
	transaction pgx.Tx,
	locationID int64,
) (Latest, error) {
	const query = `
		SELECT
			l.location_id,
			l.open_meteo_location_id,
			l.city,
			l.country,
			l.country_code,
			l.latitude,
			l.longitude,
			l.timezone,
			l.elevation,
			l.population,
			l.administrative_area,
			l.created_at,
			l.updated_at,
			f.event_id,
			f.schema_version,
			f.source,
			f.retrieved_at,
			f.timezone,
			f.current_valid_at,
			f.current_interval_seconds,
			f.current_temperature_2m,
			f.current_apparent_temperature,
			f.current_relative_humidity_2m,
			f.current_precipitation,
			f.current_precipitation_probability,
			f.current_rain,
			f.current_showers,
			f.current_snowfall,
			f.current_weather_code,
			f.current_cloud_cover,
			f.current_pressure_msl,
			f.current_visibility,
			f.current_wind_speed_10m,
			f.current_wind_direction_10m,
			f.current_wind_gusts_10m,
			f.current_uv_index,
			f.current_is_day
		FROM public.latest_operational_forecasts AS f
		JOIN public.locations AS l USING (location_id)
		WHERE f.location_id = $1
		FOR SHARE OF f
	`

	var latest Latest
	latest.Snapshot.LocationID = locationID

	err := transaction.QueryRow(ctx, query, locationID).Scan(
		&latest.Location.ID,
		&latest.Location.OpenMeteoLocationID,
		&latest.Location.City,
		&latest.Location.Country,
		&latest.Location.CountryCode,
		&latest.Location.Latitude,
		&latest.Location.Longitude,
		&latest.Location.Timezone,
		&latest.Location.Elevation,
		&latest.Location.Population,
		&latest.Location.AdministrativeArea,
		&latest.Location.CreatedAt,
		&latest.Location.UpdatedAt,
		&latest.EventID,
		&latest.SchemaVersion,
		&latest.Snapshot.Source,
		&latest.Snapshot.RetrievedAt,
		&latest.Snapshot.Timezone,
		&latest.Snapshot.Current.ValidAt,
		&latest.Snapshot.Current.IntervalSeconds,
		&latest.Snapshot.Current.Temperature2M,
		&latest.Snapshot.Current.ApparentTemperature,
		&latest.Snapshot.Current.RelativeHumidity2M,
		&latest.Snapshot.Current.Precipitation,
		&latest.Snapshot.Current.PrecipitationProbability,
		&latest.Snapshot.Current.Rain,
		&latest.Snapshot.Current.Showers,
		&latest.Snapshot.Current.Snowfall,
		&latest.Snapshot.Current.WeatherCode,
		&latest.Snapshot.Current.CloudCover,
		&latest.Snapshot.Current.PressureMSL,
		&latest.Snapshot.Current.Visibility,
		&latest.Snapshot.Current.WindSpeed10M,
		&latest.Snapshot.Current.WindDirection10M,
		&latest.Snapshot.Current.WindGusts10M,
		&latest.Snapshot.Current.UVIndex,
		&latest.Snapshot.Current.IsDay,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Latest{}, fmt.Errorf(
			"%w for location %d",
			ErrLatestForecastNotFound,
			locationID,
		)
	}

	if err != nil {
		return Latest{}, fmt.Errorf("read latest current forecast: %w", err)
	}

	return latest, nil
}

func readHourly(
	ctx context.Context,
	transaction pgx.Tx,
	locationID int64,
) ([]forecast.OperationalHourlyForecast, error) {
	const query = `
		SELECT
			valid_at,
			temperature_2m,
			apparent_temperature,
			relative_humidity_2m,
			precipitation,
			precipitation_probability,
			rain,
			showers,
			snowfall,
			weather_code,
			cloud_cover,
			pressure_msl,
			visibility,
			wind_speed_10m,
			wind_direction_10m,
			wind_gusts_10m,
			uv_index,
			is_day
		FROM public.latest_operational_hourly
		WHERE location_id = $1
		ORDER BY valid_at
	`

	rows, err := transaction.Query(ctx, query, locationID)
	if err != nil {
		return nil, fmt.Errorf("query latest hourly forecast: %w", err)
	}
	defer rows.Close()

	result := make([]forecast.OperationalHourlyForecast, 0)
	for rows.Next() {
		var item forecast.OperationalHourlyForecast

		if err := rows.Scan(
			&item.ValidAt,
			&item.Temperature2M,
			&item.ApparentTemperature,
			&item.RelativeHumidity2M,
			&item.Precipitation,
			&item.PrecipitationProbability,
			&item.Rain,
			&item.Showers,
			&item.Snowfall,
			&item.WeatherCode,
			&item.CloudCover,
			&item.PressureMSL,
			&item.Visibility,
			&item.WindSpeed10M,
			&item.WindDirection10M,
			&item.WindGusts10M,
			&item.UVIndex,
			&item.IsDay,
		); err != nil {
			return nil, fmt.Errorf("scan latest hourly forecast: %w", err)
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate latest hourly forecast: %w", err)
	}

	return result, nil
}

func readDaily(
	ctx context.Context,
	transaction pgx.Tx,
	locationID int64,
) ([]forecast.DailyForecast, error) {
	const query = `
		SELECT
			forecast_date,
			weather_code,
			temperature_2m_max,
			temperature_2m_min,
			apparent_temperature_max,
			apparent_temperature_min,
			precipitation_sum,
			precipitation_probability_max,
			precipitation_hours,
			wind_speed_10m_max,
			wind_gusts_10m_max,
			wind_direction_10m_dominant,
			sunrise,
			sunset,
			daylight_duration_seconds,
			uv_index_max
		FROM public.latest_operational_daily
		WHERE location_id = $1
		ORDER BY forecast_date
	`

	rows, err := transaction.Query(ctx, query, locationID)
	if err != nil {
		return nil, fmt.Errorf("query latest daily forecast: %w", err)
	}
	defer rows.Close()

	result := make([]forecast.DailyForecast, 0)
	for rows.Next() {
		var (
			item         forecast.DailyForecast
			forecastDate time.Time
		)

		if err := rows.Scan(
			&forecastDate,
			&item.WeatherCode,
			&item.Temperature2MMax,
			&item.Temperature2MMin,
			&item.ApparentTemperatureMax,
			&item.ApparentTemperatureMin,
			&item.PrecipitationSum,
			&item.PrecipitationProbabilityMax,
			&item.PrecipitationHours,
			&item.WindSpeed10MMax,
			&item.WindGusts10MMax,
			&item.WindDirection10MDominant,
			&item.Sunrise,
			&item.Sunset,
			&item.DaylightDurationSeconds,
			&item.UVIndexMax,
		); err != nil {
			return nil, fmt.Errorf("scan latest daily forecast: %w", err)
		}

		item.Date = forecastDate.Format(time.DateOnly)
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate latest daily forecast: %w", err)
	}

	return result, nil
}
