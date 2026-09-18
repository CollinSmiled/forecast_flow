package openmeteo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

const (
	geocodingBaseURL    = "https://geocoding-api.open-meteo.com"
	defaultHTTPTimeout  = 10 * time.Second
	maximumSearchLimit  = 100
	maximumResponseSize = 2 << 20
)

type GeocodingClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewGeocodingClient() *GeocodingClient {
	return newGeocodingClient(
		geocodingBaseURL,
		&http.Client{
			Timeout: defaultHTTPTimeout,
		},
	)
}

func newGeocodingClient(
	baseURL string,
	httpClient *http.Client,
) *GeocodingClient {
	return &GeocodingClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (client *GeocodingClient) SearchLocations(
	ctx context.Context,
	searchTerm string,
	countryCode string,
	limit int,
) ([]location.Location, error) {
	searchTerm = strings.TrimSpace(searchTerm)
	if searchTerm == "" {
		return nil, errors.New(
			"search Open-Meteo locations: search term is required",
		)
	}

	if limit < 1 || limit > maximumSearchLimit {
		return nil, fmt.Errorf(
			"search Open-Meteo locations: limit must be between 1 and %d",
			maximumSearchLimit,
		)
	}

	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if countryCode != "" && len(countryCode) != 2 {
		return nil, errors.New(
			"search Open-Meteo locations: country code must contain two characters",
		)
	}

	endpoint, err := url.Parse(client.baseURL + "/v1/search")
	if err != nil {
		return nil, fmt.Errorf("parse Open-Meteo URL: %w", err)
	}

	parameters := endpoint.Query()
	parameters.Set("name", searchTerm)
	parameters.Set("count", strconv.Itoa(limit))
	parameters.Set("format", "json")
	parameters.Set("language", "en")

	if countryCode != "" {
		parameters.Set("countryCode", countryCode)
	}

	endpoint.RawQuery = parameters.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create Open-Meteo geocoding request: %w",
			err,
		)
	}

	request.Header.Set("Accept", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"request Open-Meteo geocoding: %w",
			err,
		)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(
		response.Body,
		maximumResponseSize+1,
	))
	if err != nil {
		return nil, fmt.Errorf(
			"read Open-Meteo geocoding response: %w",
			err,
		)
	}

	if len(body) > maximumResponseSize {
		return nil, errors.New(
			"read Open-Meteo geocoding response: response is too large",
		)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, decodeGeocodingError(response.StatusCode, body)
	}

	var payload geocodingResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf(
			"decode Open-Meteo geocoding response: %w",
			err,
		)
	}

	locations := make([]location.Location, 0, len(payload.Results))

	for index, result := range payload.Results {
		if result.ID == 0 ||
			result.Name == "" ||
			result.Country == "" ||
			result.CountryCode == "" ||
			result.Timezone == "" {
			return nil, fmt.Errorf(
				"decode Open-Meteo geocoding response: result %d is incomplete",
				index,
			)
		}

		locations = append(locations, result.location())
	}

	return locations, nil
}

type geocodingResponse struct {
	Results []geocodingResult `json:"results"`
}

type geocodingResult struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Country     string   `json:"country"`
	CountryCode string   `json:"country_code"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Timezone    string   `json:"timezone"`
	Elevation   *float64 `json:"elevation"`
	Population  *int64   `json:"population"`
	Admin1      string   `json:"admin1"`
}

func (result geocodingResult) location() location.Location {
	var administrativeArea *string

	if value := strings.TrimSpace(result.Admin1); value != "" {
		administrativeArea = &value
	}

	return location.Location{
		OpenMeteoLocationID: result.ID,
		City:                result.Name,
		Country:             result.Country,
		CountryCode:         strings.ToUpper(result.CountryCode),
		Latitude:            result.Latitude,
		Longitude:           result.Longitude,
		Timezone:            result.Timezone,
		Elevation:           result.Elevation,
		Population:          result.Population,
		AdministrativeArea:  administrativeArea,
	}
}

type geocodingErrorResponse struct {
	Reason string `json:"reason"`
}

func decodeGeocodingError(statusCode int, body []byte) error {
	var payload geocodingErrorResponse

	if err := json.Unmarshal(body, &payload); err == nil &&
		payload.Reason != "" {
		return fmt.Errorf(
			"Open-Meteo geocoding returned status %d: %s",
			statusCode,
			payload.Reason,
		)
	}

	return fmt.Errorf(
		"Open-Meteo geocoding returned status %d",
		statusCode,
	)
}
