package openmeteo

import (
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
)

const operationalForecastSource = "open_meteo_best_match"

func mapOperationalForecast(
	locationID int64,
	retrievedAt time.Time,
	payload forecastResponse,
) (forecast.OperationalForecastSnapshot, error) {
	if locationID <= 0 {
		return forecast.OperationalForecastSnapshot{}, fmt.Errorf(
			"location ID must be greater than zero",
		)
	}

	if retrievedAt.IsZero() {
		return forecast.OperationalForecastSnapshot{}, fmt.Errorf(
			"retrieval time is required",
		)
	}

	timezone, err := loadOpenMeteoTimezone(payload.Timezone)
	if err != nil {
		return forecast.OperationalForecastSnapshot{}, err
	}

	current, err := mapCurrentConditions(payload.Current, timezone)
	if err != nil {
		return forecast.OperationalForecastSnapshot{}, fmt.Errorf(
			"map current conditions: %w",
			err,
		)
	}

	hourly, err := mapOperationalHourlyForecasts(
		payload.Hourly,
		timezone,
	)
	if err != nil {
		return forecast.OperationalForecastSnapshot{}, fmt.Errorf(
			"map operational hourly forecasts: %w",
			err,
		)
	}

	daily, err := mapDailyForecasts(payload)
	if err != nil {
		return forecast.OperationalForecastSnapshot{}, fmt.Errorf(
			"map operational daily forecasts: %w",
			err,
		)
	}

	return forecast.OperationalForecastSnapshot{
		LocationID:  locationID,
		Source:      operationalForecastSource,
		RetrievedAt: retrievedAt.UTC(),
		Timezone:    payload.Timezone,
		Current:     current,
		Hourly:      hourly,
		Daily:       daily,
	}, nil
}

func mapCurrentConditions(
	payload forecastCurrent,
	timezone *time.Location,
) (forecast.CurrentConditions, error) {
	validAt, err := time.ParseInLocation(
		openMeteoDateTimeLayout,
		payload.Time,
		timezone,
	)
	if err != nil {
		return forecast.CurrentConditions{}, fmt.Errorf(
			"parse current time %q: %w",
			payload.Time,
			err,
		)
	}

	isDay, err := mapIsDay(payload.IsDay)
	if err != nil {
		return forecast.CurrentConditions{}, fmt.Errorf(
			"map is_day: %w",
			err,
		)
	}

	return forecast.CurrentConditions{
		ValidAt:         validAt.UTC(),
		IntervalSeconds: payload.IntervalSeconds,
		WeatherMetrics: forecast.WeatherMetrics{
			Temperature2M:       payload.Temperature2M,
			ApparentTemperature: payload.ApparentTemperature,
			RelativeHumidity2M:  payload.RelativeHumidity2M,
			Precipitation:       payload.Precipitation,
			Rain:                payload.Rain,
			Showers:             payload.Showers,
			Snowfall:            payload.Snowfall,
			WeatherCode:         payload.WeatherCode,
			CloudCover:          payload.CloudCover,
			PressureMSL:         payload.PressureMSL,
			WindSpeed10M:        payload.WindSpeed10M,
			WindDirection10M:    payload.WindDirection10M,
			WindGusts10M:        payload.WindGusts10M,
			IsDay:               isDay,
		},
	}, nil
}

func mapOperationalHourlyForecasts(
	payload forecastHourly,
	timezone *time.Location,
) ([]forecast.OperationalHourlyForecast, error) {
	rowCount := len(payload.Time)

	if rowCount == 0 {
		return nil, fmt.Errorf("hourly forecast times are missing")
	}

	series := []seriesLength{
		{"temperature_2m", len(payload.Temperature2M)},
		{"apparent_temperature", len(payload.ApparentTemperature)},
		{"relative_humidity_2m", len(payload.RelativeHumidity2M)},
		{"precipitation", len(payload.Precipitation)},
		{
			"precipitation_probability",
			len(payload.PrecipitationProbability),
		},
		{"rain", len(payload.Rain)},
		{"showers", len(payload.Showers)},
		{"snowfall", len(payload.Snowfall)},
		{"weather_code", len(payload.WeatherCode)},
		{"cloud_cover", len(payload.CloudCover)},
		{"pressure_msl", len(payload.PressureMSL)},
		{"visibility", len(payload.Visibility)},
		{"wind_speed_10m", len(payload.WindSpeed10M)},
		{"wind_direction_10m", len(payload.WindDirection10M)},
		{"wind_gusts_10m", len(payload.WindGusts10M)},
		{"uv_index", len(payload.UVIndex)},
		{"is_day", len(payload.IsDay)},
	}

	if err := validateSeriesLengths(rowCount, series); err != nil {
		return nil, fmt.Errorf(
			"validate Open-Meteo operational hourly data: %w",
			err,
		)
	}

	result := make(
		[]forecast.OperationalHourlyForecast,
		0,
		rowCount,
	)

	for index, rawTime := range payload.Time {
		validAt, err := time.ParseInLocation(
			openMeteoDateTimeLayout,
			rawTime,
			timezone,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"parse operational hourly time at index %d: %w",
				index,
				err,
			)
		}

		isDay, err := mapIsDay(payload.IsDay[index])
		if err != nil {
			return nil, fmt.Errorf(
				"map is_day at index %d: %w",
				index,
				err,
			)
		}

		result = append(result, forecast.OperationalHourlyForecast{
			ValidAt: validAt.UTC(),
			WeatherMetrics: forecast.WeatherMetrics{
				Temperature2M:            payload.Temperature2M[index],
				ApparentTemperature:      payload.ApparentTemperature[index],
				RelativeHumidity2M:       payload.RelativeHumidity2M[index],
				Precipitation:            payload.Precipitation[index],
				PrecipitationProbability: payload.PrecipitationProbability[index],
				Rain:                     payload.Rain[index],
				Showers:                  payload.Showers[index],
				Snowfall:                 payload.Snowfall[index],
				WeatherCode:              payload.WeatherCode[index],
				CloudCover:               payload.CloudCover[index],
				PressureMSL:              payload.PressureMSL[index],
				Visibility:               payload.Visibility[index],
				WindSpeed10M:             payload.WindSpeed10M[index],
				WindDirection10M:         payload.WindDirection10M[index],
				WindGusts10M:             payload.WindGusts10M[index],
				UVIndex:                  payload.UVIndex[index],
				IsDay:                    isDay,
			},
		})
	}

	return result, nil
}
