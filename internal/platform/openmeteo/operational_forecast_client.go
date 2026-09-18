package openmeteo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type OperationalForecastClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOperationalForecastClient() *OperationalForecastClient {
	return newOperationalForecastClient(
		operationalForecastBaseURL,
		&http.Client{
			Timeout: defaultHTTPTimeout,
		},
	)
}

func newOperationalForecastClient(
	baseURL string,
	httpClient *http.Client,
) *OperationalForecastClient {
	return &OperationalForecastClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (client *OperationalForecastClient) fetch(
	ctx context.Context,
	request OperationalForecastRequest,
) (forecastResponse, error) {
	endpoint, err := buildOperationalForecastEndpoint(
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
		"forecast",
		&payload,
	); err != nil {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo operational forecast: %w",
			err,
		)
	}

	if payload.Timezone == "" {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo operational forecast: response timezone is missing",
		)
	}

	if payload.Current.Time == "" {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo operational forecast: current data is missing",
		)
	}

	if len(payload.Hourly.Time) == 0 {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo operational forecast: hourly data is missing",
		)
	}

	if len(payload.Daily.Time) == 0 {
		return forecastResponse{}, fmt.Errorf(
			"fetch Open-Meteo operational forecast: daily data is missing",
		)
	}

	return payload, nil
}
