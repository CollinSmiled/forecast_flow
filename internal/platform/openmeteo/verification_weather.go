package openmeteo

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	verificationWeatherBaseURL = "https://archive-api.open-meteo.com"
	maximumVerificationDays    = 31
)

type VerificationWeatherRequest struct {
	Latitude  float64
	Longitude float64
	Timezone  string
	StartDate time.Time
	EndDate   time.Time
}

func (request VerificationWeatherRequest) validate() error {
	if request.Latitude < -90 || request.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if request.Longitude < -180 || request.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	if strings.TrimSpace(request.Timezone) == "" {
		return fmt.Errorf("timezone is required")
	}
	if request.StartDate.IsZero() || request.EndDate.IsZero() {
		return fmt.Errorf("start and end dates are required")
	}

	startDate := dateOnlyUTC(request.StartDate)
	endDate := dateOnlyUTC(request.EndDate)
	if endDate.Before(startDate) {
		return fmt.Errorf("end date must not be before start date")
	}
	if int(endDate.Sub(startDate).Hours()/24)+1 > maximumVerificationDays {
		return fmt.Errorf(
			"verification range must not exceed %d days",
			maximumVerificationDays,
		)
	}

	return nil
}

func buildVerificationWeatherEndpoint(
	baseURL string,
	request VerificationWeatherRequest,
) (*url.URL, error) {
	if err := request.validate(); err != nil {
		return nil, fmt.Errorf("validate verification weather request: %w", err)
	}

	endpoint, err := url.Parse(strings.TrimRight(baseURL, "/") + "/v1/archive")
	if err != nil {
		return nil, fmt.Errorf("parse verification weather URL: %w", err)
	}

	parameters := endpoint.Query()
	parameters.Set("latitude", strconv.FormatFloat(request.Latitude, 'f', -1, 64))
	parameters.Set("longitude", strconv.FormatFloat(request.Longitude, 'f', -1, 64))
	parameters.Set("timezone", strings.TrimSpace(request.Timezone))
	parameters.Set("models", "best_match")
	parameters.Set("start_date", request.StartDate.Format(time.DateOnly))
	parameters.Set("end_date", request.EndDate.Format(time.DateOnly))
	parameters.Set("hourly", strings.Join(verificationHourlyVariables, ","))
	parameters.Set("temperature_unit", "celsius")
	parameters.Set("wind_speed_unit", "kmh")
	parameters.Set("precipitation_unit", "mm")
	parameters.Set("timeformat", "iso8601")
	endpoint.RawQuery = parameters.Encode()

	return endpoint, nil
}

func dateOnlyUTC(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
