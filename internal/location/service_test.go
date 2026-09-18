package location

import (
	"context"
	"errors"
	"testing"
)

type stubGeocoder struct {
	searchCountryCode string
	searchResults     []Location
	searchError       error

	requestedLocationID int64
	location            Location
	locationError       error
}

func (stub *stubGeocoder) SearchLocations(
	_ context.Context,
	_ string,
	countryCode string,
	_ int,
) ([]Location, error) {
	stub.searchCountryCode = countryCode

	return stub.searchResults, stub.searchError
}

func (stub *stubGeocoder) GetLocation(
	_ context.Context,
	openMeteoLocationID int64,
) (Location, error) {
	stub.requestedLocationID = openMeteoLocationID

	return stub.location, stub.locationError
}

type stubStore struct {
	upsertedLocation Location
	upsertResult     Location
	upsertError      error

	searchCountryCode string
	searchResults     []Location
	searchError       error
}

func (stub *stubStore) Upsert(
	_ context.Context,
	candidate Location,
) (Location, error) {
	stub.upsertedLocation = candidate

	return stub.upsertResult, stub.upsertError
}

func (stub *stubStore) Search(
	_ context.Context,
	_ string,
	countryCode string,
	_ int,
) ([]Location, error) {
	stub.searchCountryCode = countryCode

	return stub.searchResults, stub.searchError
}

func TestServiceSearchAvailable(t *testing.T) {
	geocoder := &stubGeocoder{
		searchResults: []Location{
			{
				City:        "Jakarta",
				CountryCode: "ID",
			},
			{
				City:        "Different Jakarta",
				CountryCode: "US",
			},
		},
	}
	store := &stubStore{}

	service, err := NewService(
		geocoder,
		store,
		[]string{"ID", "JP"},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	found, err := service.SearchAvailable(
		context.Background(),
		"Jakarta",
		"id",
		10,
	)
	if err != nil {
		t.Fatalf("search available locations: %v", err)
	}

	if geocoder.searchCountryCode != "ID" {
		t.Errorf(
			"country code = %q, want %q",
			geocoder.searchCountryCode,
			"ID",
		)
	}

	if len(found) != 1 {
		t.Fatalf(
			"result count = %d, want 1",
			len(found),
		)
	}

	if found[0].City != "Jakarta" {
		t.Errorf(
			"city = %q, want %q",
			found[0].City,
			"Jakarta",
		)
	}
}

func TestServiceRejectsUnsupportedCountry(t *testing.T) {
	geocoder := &stubGeocoder{}
	store := &stubStore{}

	service, err := NewService(
		geocoder,
		store,
		[]string{"ID", "JP"},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.SearchAvailable(
		context.Background(),
		"New York",
		"US",
		10,
	)
	if !errors.Is(err, ErrCountryNotAllowed) {
		t.Fatalf(
			"error = %v, want ErrCountryNotAllowed",
			err,
		)
	}
}

func TestServiceAddsAuthoritativeLocation(t *testing.T) {
	authoritative := Location{
		OpenMeteoLocationID: 1642911,
		City:                "Jakarta",
		Country:             "Indonesia",
		CountryCode:         "ID",
		Latitude:            -6.21462,
		Longitude:           106.84513,
		Timezone:            "Asia/Jakarta",
	}

	geocoder := &stubGeocoder{
		location: authoritative,
	}
	store := &stubStore{
		upsertResult: Location{
			ID:                  101,
			OpenMeteoLocationID: 1642911,
			City:                "Jakarta",
			CountryCode:         "ID",
		},
	}

	service, err := NewService(
		geocoder,
		store,
		[]string{"ID", "JP"},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	saved, err := service.Add(
		context.Background(),
		1642911,
	)
	if err != nil {
		t.Fatalf("add location: %v", err)
	}

	if geocoder.requestedLocationID != 1642911 {
		t.Errorf(
			"requested location ID = %d, want %d",
			geocoder.requestedLocationID,
			1642911,
		)
	}

	if store.upsertedLocation.City != "Jakarta" {
		t.Errorf(
			"upserted city = %q, want %q",
			store.upsertedLocation.City,
			"Jakarta",
		)
	}

	if saved.ID != 101 {
		t.Errorf(
			"saved ID = %d, want %d",
			saved.ID,
			101,
		)
	}
}

func TestServiceDoesNotSaveUnsupportedLocation(t *testing.T) {
	geocoder := &stubGeocoder{
		location: Location{
			OpenMeteoLocationID: 5128581,
			City:                "New York",
			CountryCode:         "US",
		},
	}
	store := &stubStore{}

	service, err := NewService(
		geocoder,
		store,
		[]string{"ID", "JP"},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.Add(
		context.Background(),
		5128581,
	)
	if !errors.Is(err, ErrCountryNotAllowed) {
		t.Fatalf(
			"error = %v, want ErrCountryNotAllowed",
			err,
		)
	}

	if store.upsertedLocation.OpenMeteoLocationID != 0 {
		t.Error("unsupported location was sent to the store")
	}
}

func TestServiceSearchesSavedLocations(t *testing.T) {
	geocoder := &stubGeocoder{}
	store := &stubStore{
		searchResults: []Location{
			{
				ID:          101,
				City:        "Jakarta",
				CountryCode: "ID",
			},
		},
	}

	service, err := NewService(
		geocoder,
		store,
		[]string{"ID"},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	found, err := service.SearchSaved(
		context.Background(),
		"Jakarta",
		"id",
		10,
	)
	if err != nil {
		t.Fatalf("search saved locations: %v", err)
	}

	if store.searchCountryCode != "ID" {
		t.Errorf(
			"country code = %q, want %q",
			store.searchCountryCode,
			"ID",
		)
	}

	if len(found) != 1 {
		t.Fatalf(
			"result count = %d, want 1",
			len(found),
		)
	}
}
