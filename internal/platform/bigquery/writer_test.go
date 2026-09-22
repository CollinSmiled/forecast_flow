package bigquery

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	cloudbigquery "cloud.google.com/go/bigquery"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
)

func TestNewWriterRequiresConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		client    *cloudbigquery.Client
		datasetID string
		location  string
	}{
		{name: "client", datasetID: "forecast_raw", location: "asia-southeast2"},
		{name: "dataset", client: &cloudbigquery.Client{}, location: "asia-southeast2"},
		{name: "location", client: &cloudbigquery.Client{}, datasetID: "forecast_raw"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer, err := NewWriter(test.client, test.datasetID, test.location)
			if err == nil {
				t.Fatal("NewWriter() error = nil, want an error")
			}
			if writer != nil {
				t.Fatalf("NewWriter() writer = %#v, want nil", writer)
			}
		})
	}
}

func TestWriterLoadsOperationalForecastBatchAsNDJSON(t *testing.T) {
	operational := &fakeLoader{}
	modelRuns := &fakeLoader{}
	writer := newWriter(operational, modelRuns)
	rows := []coldstore.OperationalForecastRow{
		{EventID: "event-operational-1", LocationID: 3},
		{EventID: "event-operational-2", LocationID: 4},
	}

	err := writer.AppendOperationalForecasts(context.Background(), rows)
	if err != nil {
		t.Fatalf("AppendOperationalForecasts() error = %v", err)
	}

	assertLoadedRows(t, operational, "operational", 2)
	if !strings.Contains(string(operational.data), `"event_id":"event-operational-1"`) {
		t.Fatalf("load data = %s, want snake-case event_id", operational.data)
	}
	if modelRuns.calls != 0 {
		t.Fatalf("model-run Load() calls = %d, want 0", modelRuns.calls)
	}
}

func TestWriterLoadsModelRunBatchAsNDJSON(t *testing.T) {
	operational := &fakeLoader{}
	modelRuns := &fakeLoader{}
	writer := newWriter(operational, modelRuns)
	rows := []coldstore.ModelRunRow{
		{EventID: "event-model-run-1", ModelID: "ecmwf_ifs"},
	}

	err := writer.AppendModelRuns(context.Background(), rows)
	if err != nil {
		t.Fatalf("AppendModelRuns() error = %v", err)
	}

	assertLoadedRows(t, modelRuns, "model_run", 1)
	if operational.calls != 0 {
		t.Fatalf("operational Load() calls = %d, want 0", operational.calls)
	}
}

func TestWriterSkipsEmptyBatch(t *testing.T) {
	operational := &fakeLoader{}
	writer := newWriter(operational, &fakeLoader{})

	if err := writer.AppendOperationalForecasts(context.Background(), nil); err != nil {
		t.Fatalf("AppendOperationalForecasts() error = %v", err)
	}
	if operational.calls != 0 {
		t.Fatalf("Load() calls = %d, want 0", operational.calls)
	}
}

func TestWriterUsesDeterministicJobID(t *testing.T) {
	first := &fakeLoader{}
	second := &fakeLoader{}
	rows := []coldstore.OperationalForecastRow{{EventID: "event-1"}}

	if err := newWriter(first, &fakeLoader{}).
		AppendOperationalForecasts(context.Background(), rows); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := newWriter(second, &fakeLoader{}).
		AppendOperationalForecasts(context.Background(), rows); err != nil {
		t.Fatalf("second append: %v", err)
	}
	if first.jobID != second.jobID {
		t.Fatalf("job IDs = %q and %q, want equal", first.jobID, second.jobID)
	}
}

func TestWriterWrapsLoadError(t *testing.T) {
	loadError := errors.New("load failed")
	writer := newWriter(&fakeLoader{err: loadError}, &fakeLoader{})

	err := writer.AppendOperationalForecasts(
		context.Background(),
		[]coldstore.OperationalForecastRow{{EventID: "event-1"}},
	)
	if !errors.Is(err, loadError) {
		t.Fatalf("AppendOperationalForecasts() error = %v, want %v", err, loadError)
	}
}

func assertLoadedRows(
	t *testing.T,
	loader *fakeLoader,
	wantPrefix string,
	wantRows int,
) {
	t.Helper()

	if loader.calls != 1 {
		t.Fatalf("Load() calls = %d, want 1", loader.calls)
	}
	if !strings.HasPrefix(loader.jobID, "forecast_flow_"+wantPrefix+"_") {
		t.Fatalf("job ID = %q, want %q prefix", loader.jobID, wantPrefix)
	}

	decoder := json.NewDecoder(strings.NewReader(string(loader.data)))
	rows := 0
	for decoder.More() {
		var row map[string]any
		if err := decoder.Decode(&row); err != nil {
			t.Fatalf("decode NDJSON row: %v", err)
		}
		rows++
	}
	if rows != wantRows {
		t.Fatalf("NDJSON rows = %d, want %d", rows, wantRows)
	}
}

type fakeLoader struct {
	calls int
	jobID string
	data  []byte
	err   error
}

func (f *fakeLoader) Load(_ context.Context, jobID string, data []byte) error {
	f.calls++
	f.jobID = jobID
	f.data = append([]byte(nil), data...)
	return f.err
}
