package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

const (
	defaultLocationSearchLimit = 10
	maximumLocationSearchLimit = 100
	maximumRequestBodySize     = 1 << 20
)

type LocationService interface {
	SearchAvailable(
		ctx context.Context,
		searchTerm string,
		countryCode string,
		limit int,
	) ([]location.Location, error)

	Add(
		ctx context.Context,
		openMeteoLocationID int64,
	) (location.Location, error)

	SearchSaved(
		ctx context.Context,
		searchTerm string,
		countryCode string,
		limit int,
	) ([]location.Location, error)
}

type locationHandler struct {
	service LocationService
}

func registerLocationRoutes(
	router *http.ServeMux,
	service LocationService,
) {
	handler := locationHandler{
		service: service,
	}

	router.HandleFunc(
		"GET /api/v1/locations/search",
		handler.searchAvailable,
	)
	router.HandleFunc(
		"GET /api/v1/locations",
		handler.searchSaved,
	)
	router.HandleFunc(
		"POST /api/v1/locations",
		handler.add,
	)
}

func (handler locationHandler) searchAvailable(
	response http.ResponseWriter,
	request *http.Request,
) {
	searchTerm, countryCode, limit, err :=
		parseLocationSearchParameters(request)
	if err != nil {
		writeAPIError(
			response,
			http.StatusBadRequest,
			"invalid_request",
			err.Error(),
		)

		return
	}

	found, err := handler.service.SearchAvailable(
		request.Context(),
		searchTerm,
		countryCode,
		limit,
	)
	if err != nil {
		writeLocationServiceError(response, err)

		return
	}

	writeJSON(
		response,
		http.StatusOK,
		newLocationListResponse(found),
	)
}

func (handler locationHandler) searchSaved(
	response http.ResponseWriter,
	request *http.Request,
) {
	searchTerm, countryCode, limit, err :=
		parseLocationSearchParameters(request)
	if err != nil {
		writeAPIError(
			response,
			http.StatusBadRequest,
			"invalid_request",
			err.Error(),
		)

		return
	}

	found, err := handler.service.SearchSaved(
		request.Context(),
		searchTerm,
		countryCode,
		limit,
	)
	if err != nil {
		writeLocationServiceError(response, err)

		return
	}

	writeJSON(
		response,
		http.StatusOK,
		newLocationListResponse(found),
	)
}

func (handler locationHandler) add(
	response http.ResponseWriter,
	request *http.Request,
) {
	request.Body = http.MaxBytesReader(
		response,
		request.Body,
		maximumRequestBodySize,
	)

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var input addLocationRequest

	if err := decoder.Decode(&input); err != nil {
		writeAPIError(
			response,
			http.StatusBadRequest,
			"invalid_request",
			"request body must contain valid JSON",
		)

		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeAPIError(
			response,
			http.StatusBadRequest,
			"invalid_request",
			"request body must contain one JSON object",
		)

		return
	}

	saved, err := handler.service.Add(
		request.Context(),
		input.OpenMeteoLocationID,
	)
	if err != nil {
		writeLocationServiceError(response, err)

		return
	}

	writeJSON(
		response,
		http.StatusCreated,
		locationResponseEnvelope{
			Data: newLocationResponse(saved),
		},
	)
}

func parseLocationSearchParameters(
	request *http.Request,
) (string, string, int, error) {
	query := request.URL.Query()

	searchTerm := strings.TrimSpace(query.Get("q"))
	if searchTerm == "" {
		return "", "", 0, errors.New(
			"query parameter q is required",
		)
	}

	countryCode := strings.ToUpper(
		strings.TrimSpace(query.Get("country_code")),
	)
	if countryCode == "" {
		return "", "", 0, errors.New(
			"query parameter country_code is required",
		)
	}

	limit := defaultLocationSearchLimit

	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			return "", "", 0, errors.New(
				"query parameter limit must be an integer",
			)
		}

		limit = parsed
	}

	if limit < 1 || limit > maximumLocationSearchLimit {
		return "", "", 0, fmt.Errorf(
			"query parameter limit must be between 1 and %d",
			maximumLocationSearchLimit,
		)
	}

	return searchTerm, countryCode, limit, nil
}

type addLocationRequest struct {
	OpenMeteoLocationID int64 `json:"open_meteo_location_id"`
}

type locationResponse struct {
	LocationID          int64   `json:"location_id,omitempty"`
	OpenMeteoLocationID int64   `json:"open_meteo_location_id"`
	City                string  `json:"city"`
	Country             string  `json:"country"`
	CountryCode         string  `json:"country_code"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
	Timezone            string  `json:"timezone"`

	Elevation          *float64 `json:"elevation"`
	Population         *int64   `json:"population"`
	AdministrativeArea *string  `json:"administrative_area"`
}

type locationListResponse struct {
	Data []locationResponse `json:"data"`
}

type locationResponseEnvelope struct {
	Data locationResponse `json:"data"`
}

func newLocationListResponse(
	locations []location.Location,
) locationListResponse {
	data := make([]locationResponse, 0, len(locations))

	for _, found := range locations {
		data = append(data, newLocationResponse(found))
	}

	return locationListResponse{
		Data: data,
	}
}

func newLocationResponse(found location.Location) locationResponse {
	return locationResponse{
		LocationID:          found.ID,
		OpenMeteoLocationID: found.OpenMeteoLocationID,
		City:                found.City,
		Country:             found.Country,
		CountryCode:         found.CountryCode,
		Latitude:            found.Latitude,
		Longitude:           found.Longitude,
		Timezone:            found.Timezone,
		Elevation:           found.Elevation,
		Population:          found.Population,
		AdministrativeArea:  found.AdministrativeArea,
	}
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

func writeLocationServiceError(
	response http.ResponseWriter,
	err error,
) {
	if errors.Is(err, location.ErrCountryCodeRequired) ||
		errors.Is(err, location.ErrCountryNotAllowed) ||
		errors.Is(err, location.ErrInvalidLocationID) {
		writeAPIError(
			response,
			http.StatusBadRequest,
			"invalid_request",
			err.Error(),
		)

		return
	}

	writeAPIError(
		response,
		http.StatusInternalServerError,
		"internal_error",
		"the request could not be completed",
	)
}

func writeAPIError(
	response http.ResponseWriter,
	statusCode int,
	code string,
	message string,
) {
	writeJSON(
		response,
		statusCode,
		apiErrorEnvelope{
			Error: apiError{
				Code:    code,
				Message: message,
			},
		},
	)
}

func writeJSON(
	response http.ResponseWriter,
	statusCode int,
	body any,
) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)

	if err := json.NewEncoder(response).Encode(body); err != nil {
		return
	}
}
