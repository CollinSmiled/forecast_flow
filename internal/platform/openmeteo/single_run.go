package openmeteo

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	singleRunsBaseURL = "https://single-runs-api.open-meteo.com"
)

type SingleRunRequest struct {
	Latitude      float64
	Longitude     float64
	Timezone      string
	ModelID       string
	ForecastRunAt time.Time
	ForecastDays  int
}

func (request SingleRunRequest) validate() error {
	if request.Latitude < -90 || request.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if request.Longitude < -180 || request.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	if strings.TrimSpace(request.Timezone) == "" {
		return fmt.Errorf("timezone is required")
	}

	if strings.TrimSpace(request.ModelID) == "" {
		return fmt.Errorf("model ID is required")
	}

	if request.ForecastRunAt.IsZero() {
		return fmt.Errorf("forecast run time is required")
	}

	runAt := request.ForecastRunAt.UTC()

	if runAt.Minute() != 0 ||
		runAt.Second() != 0 ||
		runAt.Nanosecond() != 0 {
		return fmt.Errorf(
			"forecast run time must align to a whole UTC hour",
		)
	}

	if request.ForecastDays < 1 ||
		request.ForecastDays > maximumForecastDays {
		return fmt.Errorf(
			"forecast days must be between 1 and %d",
			maximumForecastDays,
		)
	}

	return nil
}

func buildSingleRunEndpoint(
	baseURL string,
	request SingleRunRequest,
) (*url.URL, error) {
	if err := request.validate(); err != nil {
		return nil, fmt.Errorf("validate single-run request: %w", err)
	}

	endpoint, err := url.Parse(
		strings.TrimRight(baseURL, "/") + "/v1/forecast",
	)
	if err != nil {
		return nil, fmt.Errorf("parse single-runs URL: %w", err)
	}

	parameters := endpoint.Query()

	parameters.Set(
		"latitude",
		strconv.FormatFloat(request.Latitude, 'f', -1, 64),
	)
	parameters.Set(
		"longitude",
		strconv.FormatFloat(request.Longitude, 'f', -1, 64),
	)
	parameters.Set(
		"timezone",
		strings.TrimSpace(request.Timezone),
	)
	parameters.Set(
		"models",
		strings.TrimSpace(request.ModelID),
	)
	parameters.Set(
		"run",
		request.ForecastRunAt.UTC().Format("2006-01-02T15:04"),
	)
	parameters.Set(
		"forecast_days",
		strconv.Itoa(request.ForecastDays),
	)

	parameters.Set(
		"hourly",
		strings.Join(singleRunHourlyVariables, ","),
	)
	parameters.Set("temperature_unit", "celsius")
	parameters.Set("wind_speed_unit", "kmh")
	parameters.Set("precipitation_unit", "mm")
	parameters.Set("timeformat", "iso8601")

	endpoint.RawQuery = parameters.Encode()

	return endpoint, nil
}
