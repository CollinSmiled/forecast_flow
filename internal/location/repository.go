package location

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

const maximumSearchLimit = 100

var ErrLocationNotFound = errors.New("location not found")

type DBTX interface {
	Query(
		ctx context.Context,
		sql string,
		args ...any,
	) (pgx.Rows, error)

	QueryRow(
		ctx context.Context,
		sql string,
		args ...any,
	) pgx.Row
}

type Repository struct {
	database DBTX
}

func NewRepository(database DBTX) *Repository {
	return &Repository{
		database: database,
	}
}

func (repository *Repository) GetByID(
	ctx context.Context,
	locationID int64,
) (Location, error) {
	if locationID < 1 {
		return Location{}, errors.New(
			"get location: location ID must be greater than zero",
		)
	}

	const query = `
		SELECT
			location_id,
			open_meteo_location_id,
			city,
			country,
			country_code,
			latitude,
			longitude,
			timezone,
			elevation,
			population,
			administrative_area,
			created_at,
			updated_at
		FROM public.locations
		WHERE location_id = $1
	`

	found, err := scanLocation(
		repository.database.QueryRow(ctx, query, locationID),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Location{}, fmt.Errorf(
			"%w: %d",
			ErrLocationNotFound,
			locationID,
		)
	}

	if err != nil {
		return Location{}, fmt.Errorf("get location: %w", err)
	}

	return found, nil
}

func (repository *Repository) Upsert(
	ctx context.Context,
	candidate Location,
) (Location, error) {
	const query = `
		INSERT INTO public.locations (
			open_meteo_location_id,
			city,
			country,
			country_code,
			latitude,
			longitude,
			timezone,
			elevation,
			population,
			administrative_area
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10
		)
		ON CONFLICT (open_meteo_location_id)
		DO UPDATE SET
			city = EXCLUDED.city,
			country = EXCLUDED.country,
			country_code = EXCLUDED.country_code,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			timezone = EXCLUDED.timezone,
			elevation = EXCLUDED.elevation,
			population = EXCLUDED.population,
			administrative_area = EXCLUDED.administrative_area,
			updated_at = CURRENT_TIMESTAMP
		RETURNING
			location_id,
			open_meteo_location_id,
			city,
			country,
			country_code,
			latitude,
			longitude,
			timezone,
			elevation,
			population,
			administrative_area,
			created_at,
			updated_at
	`

	saved, err := scanLocation(repository.database.QueryRow(
		ctx,
		query,
		candidate.OpenMeteoLocationID,
		candidate.City,
		candidate.Country,
		strings.ToUpper(candidate.CountryCode),
		candidate.Latitude,
		candidate.Longitude,
		candidate.Timezone,
		candidate.Elevation,
		candidate.Population,
		candidate.AdministrativeArea,
	))
	if err != nil {
		return Location{}, fmt.Errorf("upsert location: %w", err)
	}

	return saved, nil
}

func (repository *Repository) Search(
	ctx context.Context,
	searchTerm string,
	countryCode string,
	limit int,
) ([]Location, error) {
	searchTerm = strings.TrimSpace(searchTerm)
	if searchTerm == "" {
		return nil, errors.New("search locations: search term is required")
	}

	if limit < 1 || limit > maximumSearchLimit {
		return nil, fmt.Errorf(
			"search locations: limit must be between 1 and %d",
			maximumSearchLimit,
		)
	}

	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))

	const query = `
		SELECT
			location_id,
			open_meteo_location_id,
			city,
			country,
			country_code,
			latitude,
			longitude,
			timezone,
			elevation,
			population,
			administrative_area,
			created_at,
			updated_at
		FROM public.locations
		WHERE lower(city) LIKE lower($1) || '%'
			AND ($2 = '' OR country_code = $2)
		ORDER BY
			population DESC NULLS LAST,
			city,
			location_id
		LIMIT $3
	`

	rows, err := repository.database.Query(
		ctx,
		query,
		searchTerm,
		countryCode,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}
	defer rows.Close()

	locations := make([]Location, 0)

	for rows.Next() {
		found, err := scanLocation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}

		locations = append(locations, found)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locations: %w", err)
	}

	return locations, nil
}

type rowScanner interface {
	Scan(destinations ...any) error
}

func scanLocation(scanner rowScanner) (Location, error) {
	var found Location

	err := scanner.Scan(
		&found.ID,
		&found.OpenMeteoLocationID,
		&found.City,
		&found.Country,
		&found.CountryCode,
		&found.Latitude,
		&found.Longitude,
		&found.Timezone,
		&found.Elevation,
		&found.Population,
		&found.AdministrativeArea,
		&found.CreatedAt,
		&found.UpdatedAt,
	)

	return found, err
}
