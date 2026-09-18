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
) (singleRunResponse, error) {
	endpoint, err := buildSingleRunEndpoint(
		client.baseURL,
		request,
	)
	if err != nil {
		return singleRunResponse{}, err
	}

	var payload singleRunResponse

	if err := requestJSON(
		ctx,
		client.httpClient,
		endpoint,
		"single runs",
		&payload,
	); err != nil {
		return singleRunResponse{}, fmt.Errorf(
			"fetch Open-Meteo single run: %w",
			err,
		)
	}

	if payload.Timezone == "" {
		return singleRunResponse{}, fmt.Errorf(
			"fetch Open-Meteo single run: response timezone is missing",
		)
	}

	if len(payload.Hourly.Time) == 0 {
		return singleRunResponse{}, fmt.Errorf(
			"fetch Open-Meteo single run: hourly data is missing",
		)
	}

	return payload, nil
}

type singleRunResponse struct {
	Latitude             float64         `json:"latitude"`
	Longitude            float64         `json:"longitude"`
	GenerationTimeMS     float64         `json:"generationtime_ms"`
	UTCOffsetSeconds     int             `json:"utc_offset_seconds"`
	Timezone             string          `json:"timezone"`
	TimezoneAbbreviation string          `json:"timezone_abbreviation"`
	Elevation            *float64        `json:"elevation"`
	Hourly               singleRunHourly `json:"hourly"`
	Daily                singleRunDaily  `json:"daily"`
}

type singleRunHourly struct {
	Time                     []string   `json:"time"`
	Temperature2M            []*float64 `json:"temperature_2m"`
	ApparentTemperature      []*float64 `json:"apparent_temperature"`
	RelativeHumidity2M       []*float64 `json:"relative_humidity_2m"`
	Precipitation            []*float64 `json:"precipitation"`
	PrecipitationProbability []*float64 `json:"precipitation_probability"`
	Rain                     []*float64 `json:"rain"`
	Showers                  []*float64 `json:"showers"`
	Snowfall                 []*float64 `json:"snowfall"`
	WeatherCode              []*int     `json:"weather_code"`
	CloudCover               []*float64 `json:"cloud_cover"`
	PressureMSL              []*float64 `json:"pressure_msl"`
	Visibility               []*float64 `json:"visibility"`
	WindSpeed10M             []*float64 `json:"wind_speed_10m"`
	WindDirection10M         []*float64 `json:"wind_direction_10m"`
	WindGusts10M             []*float64 `json:"wind_gusts_10m"`
	UVIndex                  []*float64 `json:"uv_index"`
	IsDay                    []*int     `json:"is_day"`
}

type singleRunDaily struct {
	Time                        []string   `json:"time"`
	WeatherCode                 []*int     `json:"weather_code"`
	Temperature2MMax            []*float64 `json:"temperature_2m_max"`
	Temperature2MMin            []*float64 `json:"temperature_2m_min"`
	ApparentTemperatureMax      []*float64 `json:"apparent_temperature_max"`
	ApparentTemperatureMin      []*float64 `json:"apparent_temperature_min"`
	PrecipitationSum            []*float64 `json:"precipitation_sum"`
	PrecipitationProbabilityMax []*float64 `json:"precipitation_probability_max"`
	PrecipitationHours          []*float64 `json:"precipitation_hours"`
	WindSpeed10MMax             []*float64 `json:"wind_speed_10m_max"`
	WindGusts10MMax             []*float64 `json:"wind_gusts_10m_max"`
	WindDirection10MDominant    []*float64 `json:"wind_direction_10m_dominant"`
	Sunrise                     []*string  `json:"sunrise"`
	Sunset                      []*string  `json:"sunset"`
	DaylightDuration            []*float64 `json:"daylight_duration"`
	UVIndexMax                  []*float64 `json:"uv_index_max"`
}
