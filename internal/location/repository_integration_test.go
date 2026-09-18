package location_test

import (
	"context"
	"os"
	"testing"

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
}
