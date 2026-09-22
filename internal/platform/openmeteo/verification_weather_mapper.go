package openmeteo

import (
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

const verificationWeatherSource = "open_meteo_historical_weather_best_match"

func mapVerificationWeather(
	locationID int64,
	retrievedAt time.Time,
	payload forecastResponse,
) (verification.Snapshot, error) {
	if locationID < 1 {
		return verification.Snapshot{}, fmt.Errorf("location ID must be greater than zero")
	}
	if retrievedAt.IsZero() {
		return verification.Snapshot{}, fmt.Errorf("retrieval time is required")
	}

	timezone, err := loadOpenMeteoTimezone(payload.Timezone)
	if err != nil {
		return verification.Snapshot{}, err
	}

	hourly, err := mapVerificationHourlyWeather(payload.Hourly, timezone)
	if err != nil {
		return verification.Snapshot{}, err
	}

	snapshot := verification.Snapshot{
		LocationID:    locationID,
		Source:        verificationWeatherSource,
		ReferenceKind: verification.ReferenceKindReanalysis,
		RetrievedAt:   retrievedAt.UTC(),
		Timezone:      payload.Timezone,
		Hourly:        hourly,
	}
	if err := snapshot.Validate(); err != nil {
		return verification.Snapshot{}, fmt.Errorf("validate verification weather: %w", err)
	}

	return snapshot, nil
}

func mapVerificationHourlyWeather(
	payload forecastHourly,
	timezone *time.Location,
) ([]verification.HourlyWeather, error) {
	rowCount := len(payload.Time)
	if rowCount == 0 {
		return nil, fmt.Errorf("verification hourly times are missing")
	}

	series := []seriesLength{
		{"temperature_2m", len(payload.Temperature2M)},
		{"apparent_temperature", len(payload.ApparentTemperature)},
		{"relative_humidity_2m", len(payload.RelativeHumidity2M)},
		{"precipitation", len(payload.Precipitation)},
		{"rain", len(payload.Rain)},
		{"snowfall", len(payload.Snowfall)},
		{"weather_code", len(payload.WeatherCode)},
		{"cloud_cover", len(payload.CloudCover)},
		{"pressure_msl", len(payload.PressureMSL)},
		{"wind_speed_10m", len(payload.WindSpeed10M)},
		{"wind_direction_10m", len(payload.WindDirection10M)},
		{"wind_gusts_10m", len(payload.WindGusts10M)},
		{"is_day", len(payload.IsDay)},
	}
	if err := validateSeriesLengths(rowCount, series); err != nil {
		return nil, fmt.Errorf("validate Open-Meteo verification hourly data: %w", err)
	}

	result := make([]verification.HourlyWeather, 0, rowCount)
	for index, rawTime := range payload.Time {
		validAt, err := time.ParseInLocation(openMeteoDateTimeLayout, rawTime, timezone)
		if err != nil {
			return nil, fmt.Errorf(
				"parse verification hourly time at index %d: %w",
				index,
				err,
			)
		}

		isDay, err := mapIsDay(payload.IsDay[index])
		if err != nil {
			return nil, fmt.Errorf("map is_day at index %d: %w", index, err)
		}

		result = append(result, verification.HourlyWeather{
			ValidAt: validAt.UTC(),
			WeatherMetrics: forecast.WeatherMetrics{
				Temperature2M:       payload.Temperature2M[index],
				ApparentTemperature: payload.ApparentTemperature[index],
				RelativeHumidity2M:  payload.RelativeHumidity2M[index],
				Precipitation:       payload.Precipitation[index],
				Rain:                payload.Rain[index],
				Snowfall:            payload.Snowfall[index],
				WeatherCode:         payload.WeatherCode[index],
				CloudCover:          payload.CloudCover[index],
				PressureMSL:         payload.PressureMSL[index],
				WindSpeed10M:        payload.WindSpeed10M[index],
				WindDirection10M:    payload.WindDirection10M[index],
				WindGusts10M:        payload.WindGusts10M[index],
				IsDay:               isDay,
			},
		})
	}

	return result, nil
}
