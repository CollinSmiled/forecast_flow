package openmeteo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeocodingClientSearchLocations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/v1/search" {
				t.Errorf(
					"path = %q, want %q",
					request.URL.Path,
					"/v1/search",
				)
			}

			query := request.URL.Query()

			if query.Get("name") != "Jakarta" {
				t.Errorf(
					"name = %q, want %q",
					query.Get("name"),
					"Jakarta",
				)
			}

			if query.Get("countryCode") != "ID" {
				t.Errorf(
					"countryCode = %q, want %q",
					query.Get("countryCode"),
					"ID",
				)
			}

			if query.Get("count") != "5" {
				t.Errorf(
					"count = %q, want %q",
					query.Get("count"),
					"5",
				)
			}

			response.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = response.Write([]byte(`{
				"results": [
					{
						"id": 1642911,
						"name": "Jakarta",
						"latitude": -6.21462,
						"longitude": 106.84513,
						"elevation": 16,
						"country_code": "ID",
						"timezone": "Asia/Jakarta",
						"population": 8540121,
						"country": "Indonesia",
						"admin1": "Jakarta"
					}
				]
			}`))
		},
	))
	defer server.Close()

	client := newGeocodingClient(
		server.URL,
		server.Client(),
	)

	locations, err := client.SearchLocations(
		context.Background(),
		" Jakarta ",
		"id",
		5,
	)
	if err != nil {
		t.Fatalf("search locations: %v", err)
	}

	if len(locations) != 1 {
		t.Fatalf(
			"location count = %d, want 1",
			len(locations),
		)
	}

	found := locations[0]

	if found.OpenMeteoLocationID != 1642911 {
		t.Errorf(
			"Open-Meteo ID = %d, want %d",
			found.OpenMeteoLocationID,
			1642911,
		)
	}

	if found.City != "Jakarta" {
		t.Errorf(
			"city = %q, want %q",
			found.City,
			"Jakarta",
		)
	}

	if found.CountryCode != "ID" {
		t.Errorf(
			"country code = %q, want %q",
			found.CountryCode,
			"ID",
		)
	}

	if found.AdministrativeArea == nil ||
		*found.AdministrativeArea != "Jakarta" {
		t.Errorf(
			"administrative area = %v, want Jakarta",
			found.AdministrativeArea,
		)
	}
}

func TestGeocodingClientReturnsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusBadRequest)

			_, _ = response.Write([]byte(`{
				"error": true,
				"reason": "Parameter count must be between 1 and 100."
			}`))
		},
	))
	defer server.Close()

	client := newGeocodingClient(
		server.URL,
		server.Client(),
	)

	_, err := client.SearchLocations(
		context.Background(),
		"Jakarta",
		"ID",
		5,
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "status 400") {
		t.Errorf(
			"error = %q, want status 400",
			err,
		)
	}
}

func TestGeocodingClientRejectsInvalidInput(t *testing.T) {
	client := NewGeocodingClient()

	tests := []struct {
		name        string
		searchTerm  string
		countryCode string
		limit       int
	}{
		{
			name:  "empty search term",
			limit: 5,
		},
		{
			name:        "invalid country code",
			searchTerm:  "Jakarta",
			countryCode: "IDN",
			limit:       5,
		},
		{
			name:       "limit too large",
			searchTerm: "Jakarta",
			limit:      101,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.SearchLocations(
				context.Background(),
				test.searchTerm,
				test.countryCode,
				test.limit,
			)

			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
