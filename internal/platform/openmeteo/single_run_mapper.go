package openmeteo

import (
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
)

const openMeteoDateTimeLayout = "2006-01-02T15:04"

type seriesLength struct {
	name   string
	length int
}

func mapHourlyForecasts(
	run forecast.Run,
	payload singleRunResponse,
) ([]forecast.HourlyForecast, error) {
	rowCount := len(payload.Hourly.Time)

	if rowCount == 0 {
		return nil, fmt.Errorf("hourly forecast times are missing")
	}

	series := []seriesLength{
		{"temperature_2m", len(payload.Hourly.Temperature2M)},
		{
			"apparent_temperature",
			len(payload.Hourly.ApparentTemperature),
		},
		{
			"relative_humidity_2m",
			len(payload.Hourly.RelativeHumidity2M),
		},
		{"precipitation", len(payload.Hourly.Precipitation)},
		{
			"precipitation_probability",
			len(payload.Hourly.PrecipitationProbability),
		},
		{"rain", len(payload.Hourly.Rain)},
		{"showers", len(payload.Hourly.Showers)},
		{"snowfall", len(payload.Hourly.Snowfall)},
		{"weather_code", len(payload.Hourly.WeatherCode)},
		{"cloud_cover", len(payload.Hourly.CloudCover)},
		{"pressure_msl", len(payload.Hourly.PressureMSL)},
		{"visibility", len(payload.Hourly.Visibility)},
		{"wind_speed_10m", len(payload.Hourly.WindSpeed10M)},
		{
			"wind_direction_10m",
			len(payload.Hourly.WindDirection10M),
		},
		{"wind_gusts_10m", len(payload.Hourly.WindGusts10M)},
		{"uv_index", len(payload.Hourly.UVIndex)},
		{"is_day", len(payload.Hourly.IsDay)},
	}

	if err := validateSeriesLengths(rowCount, series); err != nil {
		return nil, fmt.Errorf(
			"validate Open-Meteo hourly data: %w",
			err,
		)
	}

	timezone, err := time.LoadLocation(payload.Timezone)
	if err != nil {
		return nil, fmt.Errorf(
			"load Open-Meteo timezone %q: %w",
			payload.Timezone,
			err,
		)
	}

	result := make(
		[]forecast.HourlyForecast,
		0,
		rowCount,
	)

	for index, rawTime := range payload.Hourly.Time {
		validAt, err := time.ParseInLocation(
			openMeteoDateTimeLayout,
			rawTime,
			timezone,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"parse hourly time at index %d: %w",
				index,
				err,
			)
		}

		hourly, err := forecast.NewHourlyForecast(
			run,
			validAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"create hourly forecast at index %d: %w",
				index,
				err,
			)
		}

		isDay, err := mapIsDay(payload.Hourly.IsDay[index])
		if err != nil {
			return nil, fmt.Errorf(
				"map is_day at index %d: %w",
				index,
				err,
			)
		}

		hourly.Temperature2M =
			payload.Hourly.Temperature2M[index]
		hourly.ApparentTemperature =
			payload.Hourly.ApparentTemperature[index]
		hourly.RelativeHumidity2M =
			payload.Hourly.RelativeHumidity2M[index]
		hourly.Precipitation =
			payload.Hourly.Precipitation[index]
		hourly.PrecipitationProbability =
			payload.Hourly.PrecipitationProbability[index]
		hourly.Rain = payload.Hourly.Rain[index]
		hourly.Showers = payload.Hourly.Showers[index]
		hourly.Snowfall = payload.Hourly.Snowfall[index]
		hourly.WeatherCode =
			payload.Hourly.WeatherCode[index]
		hourly.CloudCover =
			payload.Hourly.CloudCover[index]
		hourly.PressureMSL =
			payload.Hourly.PressureMSL[index]
		hourly.Visibility =
			payload.Hourly.Visibility[index]
		hourly.WindSpeed10M =
			payload.Hourly.WindSpeed10M[index]
		hourly.WindDirection10M =
			payload.Hourly.WindDirection10M[index]
		hourly.WindGusts10M =
			payload.Hourly.WindGusts10M[index]
		hourly.UVIndex = payload.Hourly.UVIndex[index]
		hourly.IsDay = isDay

		result = append(result, hourly)
	}

	return result, nil
}

func validateSeriesLengths(
	expected int,
	series []seriesLength,
) error {
	for _, item := range series {
		if item.length != expected {
			return fmt.Errorf(
				"%s contains %d values; expected %d",
				item.name,
				item.length,
				expected,
			)
		}
	}

	return nil
}

func mapIsDay(value *int) (*bool, error) {
	if value == nil {
		return nil, nil
	}

	switch *value {
	case 0:
		isDay := false
		return &isDay, nil
	case 1:
		isDay := true
		return &isDay, nil
	default:
		return nil, fmt.Errorf(
			"expected 0 or 1, got %d",
			*value,
		)
	}
}
