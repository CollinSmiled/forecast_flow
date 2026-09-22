package location_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/CollinSmiled/forecast_flow/internal/platform/postgres"
)

func TestRepositoryUpsertAndSearch(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}

	ctx := context.Background()

	pool, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()

	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer func() {
		if err := transaction.Rollback(ctx); err != nil {
			t.Errorf("rollback transaction: %v", err)
		}
	}()

	repository := location.NewRepository(transaction)

	elevation := 8.0
	population := int64(1_000)
	administrativeArea := "Test Province"

	candidate := location.Location{
		OpenMeteoLocationID: 9_999_999_001,
		City:                "Repository Test City",
		Country:             "Indonesia",
		CountryCode:         "id",
		Latitude:            -6.2,
		Longitude:           106.8,
		Timezone:            "Asia/Jakarta",
		Elevation:           &elevation,
		Population:          &population,
		AdministrativeArea:  &administrativeArea,
	}

	created, err := repository.Upsert(ctx, candidate)
	if err != nil {
		t.Fatalf("upsert location: %v", err)
	}

	if created.ID == 0 {
		t.Error("location ID was not generated")
	}

	if created.CountryCode != "ID" {
		t.Errorf(
			"country code = %q, want %q",
			created.CountryCode,
			"ID",
		)
	}

	candidate.City = "Repository Test City Updated"

	updated, err := repository.Upsert(ctx, candidate)
	if err != nil {
		t.Fatalf("update location: %v", err)
	}

	if updated.ID != created.ID {
		t.Errorf(
			"updated location ID = %d, want %d",
			updated.ID,
			created.ID,
		)
	}

	loaded, err := repository.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get location by ID: %v", err)
	}

	if loaded.City != candidate.City {
		t.Errorf(
			"loaded city = %q, want %q",
			loaded.City,
			candidate.City,
		)
	}

	found, err := repository.Search(
		ctx,
		"repository",
		"ID",
		10,
	)
	if err != nil {
		t.Fatalf("search locations: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("result count = %d, want 1", len(found))
	}

	if found[0].City != candidate.City {
		t.Errorf(
			"city = %q, want %q",
			found[0].City,
			candidate.City,
		)
	}

	allLocations, err := repository.ListAll(ctx)
	if err != nil {
		t.Fatalf("list all locations: %v", err)
	}
	if !containsLocationID(allLocations, created.ID) {
		t.Fatal("saved location was absent from the reference snapshot")
	}

	missingForecast, err := repository.ListDueForForecast(
		ctx,
		time.Now().UTC(),
		100,
	)
	if err != nil {
		t.Fatalf("list location without forecast: %v", err)
	}
	if !containsLocationID(missingForecast, created.ID) {
		t.Fatal("location without forecast was not due")
	}

	retrievedAt := time.Now().UTC().Truncate(time.Second)
	if _, err := transaction.Exec(
		ctx,
		`INSERT INTO public.latest_operational_forecasts (
			location_id,
			event_id,
			schema_version,
			source,
			retrieved_at,
			timezone,
			current_valid_at,
			current_interval_seconds
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		created.ID,
		"repository-test-forecast-event",
		1,
		"test_source",
		retrievedAt,
		candidate.Timezone,
		retrievedAt,
		900,
	); err != nil {
		t.Fatalf("insert latest forecast: %v", err)
	}

	freshLocations, err := repository.ListDueForForecast(
		ctx,
		retrievedAt.Add(-time.Hour),
		100,
	)
	if err != nil {
		t.Fatalf("list locations with fresh forecast: %v", err)
	}
	if containsLocationID(freshLocations, created.ID) {
		t.Fatal("location with a fresh forecast was due")
	}

	staleLocations, err := repository.ListDueForForecast(
		ctx,
		retrievedAt.Add(time.Hour),
		100,
	)
	if err != nil {
		t.Fatalf("list locations with stale forecast: %v", err)
	}
	if !containsLocationID(staleLocations, created.ID) {
		t.Fatal("location with a stale forecast was not due")
	}
}

func containsLocationID(locations []location.Location, locationID int64) bool {
	for _, found := range locations {
		if found.ID == locationID {
			return true
		}
	}

	return false
}
