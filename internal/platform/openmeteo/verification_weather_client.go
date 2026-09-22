package openmeteo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

type VerificationWeatherClient struct {
	baseURL    string
	httpClient *http.Client
	now        func() time.Time
}

func NewVerificationWeatherClient() *VerificationWeatherClient {
	return newVerificationWeatherClient(
		verificationWeatherBaseURL,
		&http.Client{Timeout: defaultHTTPTimeout},
	)
}

func newVerificationWeatherClient(
	baseURL string,
	httpClient *http.Client,
) *VerificationWeatherClient {
	return &VerificationWeatherClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		now:        time.Now,
	}
}

func (client *VerificationWeatherClient) FetchVerificationWeather(
	ctx context.Context,
	selectedLocation location.Location,
	startDate time.Time,
	endDate time.Time,
) (verification.Snapshot, error) {
	payload, err := client.fetch(ctx, VerificationWeatherRequest{
		Latitude:  selectedLocation.Latitude,
		Longitude: selectedLocation.Longitude,
		Timezone:  selectedLocation.Timezone,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return verification.Snapshot{}, err
	}

	snapshot, err := mapVerificationWeather(selectedLocation.ID, client.now(), payload)
	if err != nil {
		return verification.Snapshot{}, fmt.Errorf(
			"map Open-Meteo verification weather: %w",
			err,
		)
	}

	return snapshot, nil
}

func (client *VerificationWeatherClient) fetch(
	ctx context.Context,
	request VerificationWeatherRequest,
) (forecastResponse, error) {
	endpoint, err := buildVerificationWeatherEndpoint(client.baseURL, request)
	if err != nil {
		return forecastResponse{}, err
	}

	var payload forecastResponse
	if err := requestJSON(
		ctx,
		client.httpClient,
		endpoint,
		"historical weather",
		&payload,
	); err != nil {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo verification weather: %w",
			err,
		)
	}

	if strings.TrimSpace(payload.Timezone) == "" {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo verification weather: response timezone is missing",
		)
	}
	if len(payload.Hourly.Time) == 0 {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo verification weather: hourly data is missing",
		)
	}

	return payload, nil
}
