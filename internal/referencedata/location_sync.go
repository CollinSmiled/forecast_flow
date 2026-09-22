package referencedata

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

type LocationRow struct {
	LocationID          int64     `bigquery:"location_id" json:"location_id"`
	OpenMeteoLocationID int64     `bigquery:"open_meteo_location_id" json:"open_meteo_location_id"`
	City                string    `bigquery:"city" json:"city"`
	Country             string    `bigquery:"country" json:"country"`
	CountryCode         string    `bigquery:"country_code" json:"country_code"`
	Latitude            float64   `bigquery:"latitude" json:"latitude"`
	Longitude           float64   `bigquery:"longitude" json:"longitude"`
	Timezone            string    `bigquery:"timezone" json:"timezone"`
	Elevation           *float64  `bigquery:"elevation" json:"elevation"`
	Population          *int64    `bigquery:"population" json:"population"`
	AdministrativeArea  *string   `bigquery:"administrative_area" json:"administrative_area"`
	SourceCreatedAt     time.Time `bigquery:"source_created_at" json:"source_created_at"`
	SourceUpdatedAt     time.Time `bigquery:"source_updated_at" json:"source_updated_at"`
	SyncedAt            time.Time `bigquery:"synced_at" json:"synced_at"`
}

type LocationSource interface {
	ListAll(ctx context.Context) ([]location.Location, error)
}

type LocationWriter interface {
	ReplaceLocations(ctx context.Context, rows []LocationRow) error
}

type LocationSync struct {
	source LocationSource
	writer LocationWriter
	now    func() time.Time
}

func NewLocationSync(
	source LocationSource,
	writer LocationWriter,
) (*LocationSync, error) {
	if source == nil {
		return nil, errors.New("location source is required")
	}
	if writer == nil {
		return nil, errors.New("location writer is required")
	}

	return &LocationSync{
		source: source,
		writer: writer,
		now:    time.Now,
	}, nil
}

func (sync *LocationSync) Run(ctx context.Context) (int, error) {
	locations, err := sync.source.ListAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("load PostgreSQL locations: %w", err)
	}
	if len(locations) == 0 {
		return 0, errors.New("location snapshot is empty")
	}

	rows, err := mapLocationRows(locations, sync.now().UTC())
	if err != nil {
		return 0, err
	}
	if err := sync.writer.ReplaceLocations(ctx, rows); err != nil {
		return 0, fmt.Errorf("replace BigQuery locations: %w", err)
	}

	return len(rows), nil
}

func mapLocationRows(
	locations []location.Location,
	syncedAt time.Time,
) ([]LocationRow, error) {
	if syncedAt.IsZero() {
		return nil, errors.New("location synchronization time is required")
	}

	rows := make([]LocationRow, 0, len(locations))
	for index, source := range locations {
		if source.ID < 1 || source.OpenMeteoLocationID < 1 {
			return nil, fmt.Errorf("location at index %d has invalid identifiers", index)
		}
		if strings.TrimSpace(source.City) == "" ||
			strings.TrimSpace(source.Country) == "" ||
			strings.TrimSpace(source.CountryCode) == "" ||
			strings.TrimSpace(source.Timezone) == "" {
			return nil, fmt.Errorf("location at index %d has incomplete descriptive data", index)
		}
		if source.CreatedAt.IsZero() || source.UpdatedAt.IsZero() {
			return nil, fmt.Errorf("location at index %d has incomplete source timestamps", index)
		}

		rows = append(rows, LocationRow{
			LocationID:          source.ID,
			OpenMeteoLocationID: source.OpenMeteoLocationID,
			City:                source.City,
			Country:             source.Country,
			CountryCode:         strings.ToUpper(source.CountryCode),
			Latitude:            source.Latitude,
			Longitude:           source.Longitude,
			Timezone:            source.Timezone,
			Elevation:           source.Elevation,
			Population:          source.Population,
			AdministrativeArea:  source.AdministrativeArea,
			SourceCreatedAt:     source.CreatedAt.UTC(),
			SourceUpdatedAt:     source.UpdatedAt.UTC(),
			SyncedAt:            syncedAt.UTC(),
		})
	}

	return rows, nil
}
