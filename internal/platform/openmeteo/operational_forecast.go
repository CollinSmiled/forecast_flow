package openmeteo

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const operationalForecastBaseURL = "https://api.open-meteo.com"

type OperationalForecastRequest struct {
	Latitude     float64
	Longitude    float64
	Timezone     string
	ForecastDays int
}

func (request OperationalForecastRequest) validate() error {
	if request.Latitude < -90 || request.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if request.Longitude < -180 || request.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	if strings.TrimSpace(request.Timezone) == "" {
		return fmt.Errorf("timezone is required")
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

func buildOperationalForecastEndpoint(
	baseURL string,
	request OperationalForecastRequest,
) (*url.URL, error) {
	if err := request.validate(); err != nil {
		return nil, fmt.Errorf(
			"validate operational forecast request: %w",
			err,
		)
	}

	endpoint, err := url.Parse(
		strings.TrimRight(baseURL, "/") + "/v1/forecast",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse operational forecast URL: %w",
			err,
		)
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
	parameters.Set("timezone", strings.TrimSpace(request.Timezone))
	parameters.Set("models", "best_match")
	parameters.Set(
		"forecast_days",
		strconv.Itoa(request.ForecastDays),
	)
	parameters.Set(
		"current",
		strings.Join(currentForecastVariables, ","),
	)
	parameters.Set(
		"hourly",
		strings.Join(operationalHourlyVariables, ","),
	)
	parameters.Set(
		"daily",
		strings.Join(dailyForecastVariables, ","),
	)
	parameters.Set("temperature_unit", "celsius")
	parameters.Set("wind_speed_unit", "kmh")
	parameters.Set("precipitation_unit", "mm")
	parameters.Set("timeformat", "iso8601")

	endpoint.RawQuery = parameters.Encode()

	return endpoint, nil
}
