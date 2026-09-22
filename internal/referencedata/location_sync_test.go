package referencedata

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestLocationSyncReplacesSnapshot(t *testing.T) {
	instant := time.Date(2026, time.September, 22, 3, 0, 0, 0, time.UTC)
	source := &fakeLocationSource{locations: []location.Location{{
		ID:                  3,
		OpenMeteoLocationID: 1_642_911,
		City:                "Jakarta",
		Country:             "Indonesia",
		CountryCode:         "id",
		Latitude:            -6.2,
		Longitude:           106.8,
		Timezone:            "Asia/Jakarta",
		CreatedAt:           instant.Add(-time.Hour),
		UpdatedAt:           instant,
	}}}
	writer := &fakeLocationWriter{}
	sync, err := NewLocationSync(source, writer)
	if err != nil {
		t.Fatalf("create location sync: %v", err)
	}
	sync.now = func() time.Time { return instant.Add(time.Minute) }

	count, err := sync.Run(context.Background())
	if err != nil {
		t.Fatalf("run location sync: %v", err)
	}
	if count != 1 || len(writer.rows) != 1 {
		t.Fatalf("synced/written locations = %d/%d, want 1/1", count, len(writer.rows))
	}
	if writer.rows[0].CountryCode != "ID" {
		t.Errorf("country code = %q, want ID", writer.rows[0].CountryCode)
	}
	if !writer.rows[0].SyncedAt.Equal(instant.Add(time.Minute)) {
		t.Errorf("synced at = %v, want deterministic clock", writer.rows[0].SyncedAt)
	}
}

func TestLocationSyncRejectsEmptySnapshot(t *testing.T) {
	sync, err := NewLocationSync(&fakeLocationSource{}, &fakeLocationWriter{})
	if err != nil {
		t.Fatalf("create location sync: %v", err)
	}

	if _, err := sync.Run(context.Background()); err == nil {
		t.Fatal("run empty location sync error = nil, want an error")
	}
}

func TestLocationSyncReturnsSourceAndWriterErrors(t *testing.T) {
	sourceError := errors.New("source unavailable")
	sync, err := NewLocationSync(
		&fakeLocationSource{err: sourceError},
		&fakeLocationWriter{},
	)
	if err != nil {
		t.Fatalf("create source-error sync: %v", err)
	}
	if _, err := sync.Run(context.Background()); !errors.Is(err, sourceError) {
		t.Fatalf("source error = %v, want %v", err, sourceError)
	}

	instant := time.Now().UTC()
	writerError := errors.New("warehouse unavailable")
	sync, err = NewLocationSync(
		&fakeLocationSource{locations: []location.Location{{
			ID:                  1,
			OpenMeteoLocationID: 2,
			City:                "Jakarta",
			Country:             "Indonesia",
			CountryCode:         "ID",
			Timezone:            "Asia/Jakarta",
			CreatedAt:           instant,
			UpdatedAt:           instant,
		}}},
		&fakeLocationWriter{err: writerError},
	)
	if err != nil {
		t.Fatalf("create writer-error sync: %v", err)
	}
	if _, err := sync.Run(context.Background()); !errors.Is(err, writerError) {
		t.Fatalf("writer error = %v, want %v", err, writerError)
	}
}

type fakeLocationSource struct {
	locations []location.Location
	err       error
}

func (source *fakeLocationSource) ListAll(
	_ context.Context,
) ([]location.Location, error) {
	return source.locations, source.err
}

type fakeLocationWriter struct {
	rows []LocationRow
	err  error
}

func (writer *fakeLocationWriter) ReplaceLocations(
	_ context.Context,
	rows []LocationRow,
) error {
	writer.rows = append([]LocationRow(nil), rows...)
	return writer.err
}
