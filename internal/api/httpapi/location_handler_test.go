package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type stubLocationService struct {
	searchTerm        string
	countryCode       string
	limit             int
	availableResults  []location.Location
	savedResults      []location.Location
	requestedLocation int64
	addResult         location.Location
}

func (stub *stubLocationService) SearchAvailable(
	_ context.Context,
	searchTerm string,
	countryCode string,
	limit int,
) ([]location.Location, error) {
	stub.searchTerm = searchTerm
	stub.countryCode = countryCode
	stub.limit = limit

	return stub.availableResults, nil
}

func (stub *stubLocationService) Add(
	_ context.Context,
	openMeteoLocationID int64,
) (location.Location, error) {
	stub.requestedLocation = openMeteoLocationID

	return stub.addResult, nil
}

func (stub *stubLocationService) SearchSaved(
	_ context.Context,
	searchTerm string,
	countryCode string,
	limit int,
) ([]location.Location, error) {
	stub.searchTerm = searchTerm
	stub.countryCode = countryCode
	stub.limit = limit

	return stub.savedResults, nil
}

func TestSearchAvailableLocations(t *testing.T) {
	service := &stubLocationService{
		availableResults: []location.Location{
			{
				OpenMeteoLocationID: 1642911,
				City:                "Jakarta",
				Country:             "Indonesia",
				CountryCode:         "ID",
				Latitude:            -6.21462,
				Longitude:           106.84513,
				Timezone:            "Asia/Jakarta",
			},
		},
	}

	router := http.NewServeMux()
	registerLocationRoutes(router, service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/locations/search?q=Jakarta&country_code=id&limit=5",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	if service.searchTerm != "Jakarta" {
		t.Errorf(
			"search term = %q, want %q",
			service.searchTerm,
			"Jakarta",
		)
	}

	if service.countryCode != "ID" {
		t.Errorf(
			"country code = %q, want %q",
			service.countryCode,
			"ID",
		)
	}

	if service.limit != 5 {
		t.Errorf(
			"limit = %d, want %d",
			service.limit,
			5,
		)
	}

	var body locationListResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(body.Data) != 1 {
		t.Fatalf(
			"result count = %d, want 1",
			len(body.Data),
		)
	}

	if body.Data[0].City != "Jakarta" {
		t.Errorf(
			"city = %q, want %q",
			body.Data[0].City,
			"Jakarta",
		)
	}
}

func TestAddLocation(t *testing.T) {
	service := &stubLocationService{
		addResult: location.Location{
			ID:                  101,
			OpenMeteoLocationID: 1642911,
			City:                "Jakarta",
			Country:             "Indonesia",
			CountryCode:         "ID",
			Latitude:            -6.21462,
			Longitude:           106.84513,
			Timezone:            "Asia/Jakarta",
		},
	}

	router := http.NewServeMux()
	registerLocationRoutes(router, service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/locations",
		strings.NewReader(`{
			"open_meteo_location_id": 1642911
		}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusCreated,
		)
	}

	if service.requestedLocation != 1642911 {
		t.Errorf(
			"requested location ID = %d, want %d",
			service.requestedLocation,
			1642911,
		)
	}

	var body locationResponseEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Data.LocationID != 101 {
		t.Errorf(
			"location ID = %d, want %d",
			body.Data.LocationID,
			101,
		)
	}
}

func TestSearchSavedLocations(t *testing.T) {
	service := &stubLocationService{
		savedResults: []location.Location{
			{
				ID:          101,
				City:        "Jakarta",
				CountryCode: "ID",
			},
		},
	}

	router := http.NewServeMux()
	registerLocationRoutes(router, service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/locations?q=Jakarta&country_code=ID",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	if service.limit != defaultLocationSearchLimit {
		t.Errorf(
			"limit = %d, want %d",
			service.limit,
			defaultLocationSearchLimit,
		)
	}
}

func TestLocationSearchRequiresQuery(t *testing.T) {
	service := &stubLocationService{}

	router := http.NewServeMux()
	registerLocationRoutes(router, service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/locations/search?country_code=ID",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusBadRequest,
		)
	}

	var body apiErrorEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Error.Code != "invalid_request" {
		t.Errorf(
			"error code = %q, want %q",
			body.Error.Code,
			"invalid_request",
		)
	}
}
