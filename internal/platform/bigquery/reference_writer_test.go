package bigquery

import (
	"context"
	"errors"
	"strings"
	"testing"

	cloudbigquery "cloud.google.com/go/bigquery"

	"github.com/CollinSmiled/forecast_flow/internal/referencedata"
)

func TestNewReferenceWriterRequiresConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		client    *cloudbigquery.Client
		datasetID string
		location  string
	}{
		{name: "client", datasetID: "forecast_reference", location: "asia-southeast2"},
		{name: "dataset", client: &cloudbigquery.Client{}, location: "asia-southeast2"},
		{name: "location", client: &cloudbigquery.Client{}, datasetID: "forecast_reference"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer, err := NewReferenceWriter(test.client, test.datasetID, test.location)
			if err == nil {
				t.Fatal("NewReferenceWriter() error = nil, want an error")
			}
			if writer != nil {
				t.Fatalf("NewReferenceWriter() writer = %#v, want nil", writer)
			}
		})
	}
}

func TestReferenceWriterLoadsLocationSnapshotAsNDJSON(t *testing.T) {
	locations := &fakeLoader{}
	writer := newReferenceWriter(locations)

	err := writer.ReplaceLocations(
		context.Background(),
		[]referencedata.LocationRow{{
			LocationID: 3,
			City:       "Jakarta",
		}},
	)
	if err != nil {
		t.Fatalf("ReplaceLocations() error = %v", err)
	}

	assertLoadedRows(t, locations, "locations", 1)
	if !strings.Contains(string(locations.data), `"city":"Jakarta"`) {
		t.Fatalf("load data = %s, want city", locations.data)
	}
}

func TestReferenceWriterRejectsEmptySnapshot(t *testing.T) {
	locations := &fakeLoader{}
	writer := newReferenceWriter(locations)

	if err := writer.ReplaceLocations(context.Background(), nil); err == nil {
		t.Fatal("ReplaceLocations() error = nil, want an error")
	}
	if locations.calls != 0 {
		t.Fatalf("Load() calls = %d, want 0", locations.calls)
	}
}

func TestReferenceWriterWrapsLoadError(t *testing.T) {
	loadError := errors.New("load failed")
	writer := newReferenceWriter(&fakeLoader{err: loadError})

	err := writer.ReplaceLocations(
		context.Background(),
		[]referencedata.LocationRow{{LocationID: 1}},
	)
	if !errors.Is(err, loadError) {
		t.Fatalf("ReplaceLocations() error = %v, want %v", err, loadError)
	}
}
