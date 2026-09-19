package openmeteo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type SingleRunClient struct {
	baseURL    string
	httpClient *http.Client
	now        func() time.Time
}

func NewSingleRunClient() *SingleRunClient {
	return newSingleRunClient(
		singleRunsBaseURL,
		&http.Client{
			Timeout: defaultHTTPTimeout,
		},
	)
}

func newSingleRunClient(
	baseURL string,
	httpClient *http.Client,
) *SingleRunClient {
	return &SingleRunClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		now:        time.Now,
	}
}

func (client *SingleRunClient) FetchForecastRun(
	ctx context.Context,
	selectedLocation location.Location,
	model forecast.Model,
	forecastRunAt time.Time,
	forecastDays int,
) (forecast.ForecastRun, error) {
	request := SingleRunRequest{
		Latitude:      selectedLocation.Latitude,
		Longitude:     selectedLocation.Longitude,
		Timezone:      selectedLocation.Timezone,
		ModelID:       model.ID,
		ForecastRunAt: forecastRunAt,
		ForecastDays:  forecastDays,
	}

	payload, err := client.fetch(ctx, request)
	if err != nil {
		return forecast.ForecastRun{}, err
	}

	run, err := forecast.NewRun(
		selectedLocation.ID,
		model,
		forecastRunAt,
		client.now().UTC(),
	)
	if err != nil {
		return forecast.ForecastRun{}, fmt.Errorf(
			"create forecast run: %w",
			err,
		)
	}

	hourly, err := mapHourlyForecasts(run, payload)
	if err != nil {
		return forecast.ForecastRun{}, fmt.Errorf(
			"map hourly forecasts: %w",
			err,
		)
	}

	return forecast.ForecastRun{
		Run:    run,
		Hourly: hourly,
	}, nil
}

func (client *SingleRunClient) fetch(
	ctx context.Context,
	request SingleRunRequest,
) (forecastResponse, error) {
	endpoint, err := buildSingleRunEndpoint(
		client.baseURL,
		request,
	)
	if err != nil {
		return forecastResponse{}, err
	}

	var payload forecastResponse

	if err := requestJSON(
		ctx,
		client.httpClient,
		endpoint,
		"single runs",
		&payload,
	); err != nil {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo single run: %w",
			err,
		)
	}

	if payload.Timezone == "" {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo single run: response timezone is missing",
		)
	}

	if len(payload.Hourly.Time) == 0 {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo single run: hourly data is missing",
		)
	}

	return payload, nil
}
