package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
	"github.com/CollinSmiled/forecast_flow/internal/hotforecast"
)

type LatestForecastReader interface {
	GetLatest(ctx context.Context, locationID int64) (hotforecast.Latest, error)
}

type forecastHandler struct {
	reader LatestForecastReader
}

func registerForecastRoutes(
	router *http.ServeMux,
	reader LatestForecastReader,
) {
	handler := forecastHandler{reader: reader}

	router.HandleFunc(
		"GET /api/v1/locations/{location_id}/forecast",
		handler.getLatest,
	)
}

func (handler forecastHandler) getLatest(
	response http.ResponseWriter,
	request *http.Request,
) {
	locationID, err := strconv.ParseInt(
		request.PathValue("location_id"),
		10,
		64,
	)
	if err != nil || locationID < 1 {
		writeAPIError(
			response,
			http.StatusBadRequest,
			"invalid_request",
			"location_id must be a positive integer",
		)

		return
	}

	latest, err := handler.reader.GetLatest(
		request.Context(),
		locationID,
	)
	if errors.Is(err, hotforecast.ErrLatestForecastNotFound) {
		writeAPIError(
			response,
			http.StatusNotFound,
			"forecast_not_found",
			"no forecast is available for this location",
		)

		return
	}

	if err != nil {
		writeAPIError(
			response,
			http.StatusInternalServerError,
			"internal_error",
			"the request could not be completed",
		)

		return
	}

	writeJSON(
		response,
		http.StatusOK,
		latestForecastEnvelope{
			Data: newLatestForecastResponse(latest),
		},
	)
}

type latestForecastEnvelope struct {
	Data latestForecastResponse `json:"data"`
}

type latestForecastResponse struct {
	ForecastID  string                   `json:"forecast_id"`
	Location    locationResponse         `json:"location"`
	Source      string                   `json:"source"`
	RetrievedAt time.Time                `json:"retrieved_at"`
	Timezone    string                   `json:"timezone"`
	Units       forecastUnitsResponse    `json:"units"`
	Current     currentForecastResponse  `json:"current"`
	Hourly      []hourlyForecastResponse `json:"hourly"`
	Daily       []dailyForecastResponse  `json:"daily"`
}

type forecastUnitsResponse struct {
	Temperature   string `json:"temperature"`
	Precipitation string `json:"precipitation"`
	WindSpeed     string `json:"wind_speed"`
	Pressure      string `json:"pressure"`
	Visibility    string `json:"visibility"`
}

type currentForecastResponse struct {
	ValidAt         time.Time `json:"valid_at"`
	IntervalSeconds int       `json:"interval_seconds"`
	weatherResponse
}

type hourlyForecastResponse struct {
	ValidAt time.Time `json:"valid_at"`
	weatherResponse
}

type weatherResponse struct {
	Temperature2M            *float64 `json:"temperature_2m"`
	ApparentTemperature      *float64 `json:"apparent_temperature"`
	RelativeHumidity2M       *float64 `json:"relative_humidity_2m"`
	Precipitation            *float64 `json:"precipitation"`
	PrecipitationProbability *float64 `json:"precipitation_probability"`
	Rain                     *float64 `json:"rain"`
	Showers                  *float64 `json:"showers"`
	Snowfall                 *float64 `json:"snowfall"`
	WeatherCode              *int     `json:"weather_code"`
	CloudCover               *float64 `json:"cloud_cover"`
	PressureMSL              *float64 `json:"pressure_msl"`
	Visibility               *float64 `json:"visibility"`
	WindSpeed10M             *float64 `json:"wind_speed_10m"`
	WindDirection10M         *float64 `json:"wind_direction_10m"`
	WindGusts10M             *float64 `json:"wind_gusts_10m"`
	UVIndex                  *float64 `json:"uv_index"`
	IsDay                    *bool    `json:"is_day"`
}

type dailyForecastResponse struct {
	Date                        string     `json:"date"`
	WeatherCode                 *int       `json:"weather_code"`
	Temperature2MMax            *float64   `json:"temperature_2m_max"`
	Temperature2MMin            *float64   `json:"temperature_2m_min"`
	ApparentTemperatureMax      *float64   `json:"apparent_temperature_max"`
	ApparentTemperatureMin      *float64   `json:"apparent_temperature_min"`
	PrecipitationSum            *float64   `json:"precipitation_sum"`
	PrecipitationProbabilityMax *float64   `json:"precipitation_probability_max"`
	PrecipitationHours          *float64   `json:"precipitation_hours"`
	WindSpeed10MMax             *float64   `json:"wind_speed_10m_max"`
	WindGusts10MMax             *float64   `json:"wind_gusts_10m_max"`
	WindDirection10MDominant    *float64   `json:"wind_direction_10m_dominant"`
	Sunrise                     *time.Time `json:"sunrise"`
	Sunset                      *time.Time `json:"sunset"`
	DaylightDurationSeconds     *float64   `json:"daylight_duration_seconds"`
	UVIndexMax                  *float64   `json:"uv_index_max"`
}

func newLatestForecastResponse(
	latest hotforecast.Latest,
) latestForecastResponse {
	return latestForecastResponse{
		ForecastID:  latest.EventID,
		Location:    newLocationResponse(latest.Location),
		Source:      latest.Snapshot.Source,
		RetrievedAt: latest.Snapshot.RetrievedAt,
		Timezone:    latest.Snapshot.Timezone,
		Units: forecastUnitsResponse{
			Temperature:   "celsius",
			Precipitation: "mm",
			WindSpeed:     "km/h",
			Pressure:      "hPa",
			Visibility:    "m",
		},
		Current: currentForecastResponse{
			ValidAt:         latest.Snapshot.Current.ValidAt,
			IntervalSeconds: latest.Snapshot.Current.IntervalSeconds,
			weatherResponse: mapWeatherResponse(
				latest.Snapshot.Current.WeatherMetrics,
			),
		},
		Hourly: mapHourlyResponse(latest.Snapshot.Hourly),
		Daily:  mapDailyResponse(latest.Snapshot.Daily),
	}
}

func mapWeatherResponse(metrics forecast.WeatherMetrics) weatherResponse {
	return weatherResponse{
		Temperature2M:            metrics.Temperature2M,
		ApparentTemperature:      metrics.ApparentTemperature,
		RelativeHumidity2M:       metrics.RelativeHumidity2M,
		Precipitation:            metrics.Precipitation,
		PrecipitationProbability: metrics.PrecipitationProbability,
		Rain:                     metrics.Rain,
		Showers:                  metrics.Showers,
		Snowfall:                 metrics.Snowfall,
		WeatherCode:              metrics.WeatherCode,
		CloudCover:               metrics.CloudCover,
		PressureMSL:              metrics.PressureMSL,
		Visibility:               metrics.Visibility,
		WindSpeed10M:             metrics.WindSpeed10M,
		WindDirection10M:         metrics.WindDirection10M,
		WindGusts10M:             metrics.WindGusts10M,
		UVIndex:                  metrics.UVIndex,
		IsDay:                    metrics.IsDay,
	}
}

func mapHourlyResponse(
	hourly []forecast.OperationalHourlyForecast,
) []hourlyForecastResponse {
	result := make([]hourlyForecastResponse, 0, len(hourly))

	for _, item := range hourly {
		result = append(result, hourlyForecastResponse{
			ValidAt:         item.ValidAt,
			weatherResponse: mapWeatherResponse(item.WeatherMetrics),
		})
	}

	return result
}

func mapDailyResponse(
	daily []forecast.DailyForecast,
) []dailyForecastResponse {
	result := make([]dailyForecastResponse, 0, len(daily))

	for _, item := range daily {
		result = append(result, dailyForecastResponse{
			Date:                        item.Date,
			WeatherCode:                 item.WeatherCode,
			Temperature2MMax:            item.Temperature2MMax,
			Temperature2MMin:            item.Temperature2MMin,
			ApparentTemperatureMax:      item.ApparentTemperatureMax,
			ApparentTemperatureMin:      item.ApparentTemperatureMin,
			PrecipitationSum:            item.PrecipitationSum,
			PrecipitationProbabilityMax: item.PrecipitationProbabilityMax,
			PrecipitationHours:          item.PrecipitationHours,
			WindSpeed10MMax:             item.WindSpeed10MMax,
			WindGusts10MMax:             item.WindGusts10MMax,
			WindDirection10MDominant:    item.WindDirection10MDominant,
			Sunrise:                     item.Sunrise,
			Sunset:                      item.Sunset,
			DaylightDurationSeconds:     item.DaylightDurationSeconds,
			UVIndexMax:                  item.UVIndexMax,
		})
	}

	return result
}
