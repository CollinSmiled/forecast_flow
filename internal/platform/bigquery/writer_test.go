package bigquery

import (
	"context"
	"errors"
	"testing"

	cloudbigquery "cloud.google.com/go/bigquery"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
)

func TestNewWriterRequiresClient(t *testing.T) {
	writer, err := NewWriter(nil, "forecast_raw")
	if err == nil {
		t.Fatal("NewWriter() error = nil, want an error")
	}
	if writer != nil {
		t.Fatalf("NewWriter() writer = %#v, want nil", writer)
	}
}

func TestNewWriterRequiresDatasetID(t *testing.T) {
	writer, err := NewWriter(&cloudbigquery.Client{}, "  ")
	if err == nil {
		t.Fatal("NewWriter() error = nil, want an error")
	}
	if writer != nil {
		t.Fatalf("NewWriter() writer = %#v, want nil", writer)
	}
}

func TestWriterAppendsOperationalForecastWithEventInsertID(t *testing.T) {
	operational := &fakeInserter{}
	modelRuns := &fakeInserter{}
	writer := newWriter(operational, modelRuns)
	row := coldstore.OperationalForecastRow{EventID: "event-operational-1"}

	err := writer.AppendOperationalForecast(context.Background(), row)
	if err != nil {
		t.Fatalf("AppendOperationalForecast() error = %v", err)
	}

	assertSavedRow(t, operational, row.EventID, row)
	if modelRuns.calls != 0 {
		t.Fatalf("model-run Put() calls = %d, want 0", modelRuns.calls)
	}
}

func TestWriterAppendsModelRunWithEventInsertID(t *testing.T) {
	operational := &fakeInserter{}
	modelRuns := &fakeInserter{}
	writer := newWriter(operational, modelRuns)
	row := coldstore.ModelRunRow{EventID: "event-model-run-1"}

	err := writer.AppendModelRun(context.Background(), row)
	if err != nil {
		t.Fatalf("AppendModelRun() error = %v", err)
	}

	assertSavedRow(t, modelRuns, row.EventID, row)
	if operational.calls != 0 {
		t.Fatalf("operational Put() calls = %d, want 0", operational.calls)
	}
}

func TestWriterWrapsInsertError(t *testing.T) {
	insertErr := errors.New("insert failed")
	writer := newWriter(
		&fakeInserter{err: insertErr},
		&fakeInserter{},
	)

	err := writer.AppendOperationalForecast(
		context.Background(),
		coldstore.OperationalForecastRow{EventID: "event-1"},
	)
	if !errors.Is(err, insertErr) {
		t.Fatalf("AppendOperationalForecast() error = %v, want %v", err, insertErr)
	}
}

func assertSavedRow(
	t *testing.T,
	inserter *fakeInserter,
	wantInsertID string,
	wantRow interface{},
) {
	t.Helper()

	if inserter.calls != 1 {
		t.Fatalf("Put() calls = %d, want 1", inserter.calls)
	}

	saver, ok := inserter.src.(*cloudbigquery.StructSaver)
	if !ok {
		t.Fatalf("Put() source type = %T, want *bigquery.StructSaver", inserter.src)
	}
	if saver.InsertID != wantInsertID {
		t.Fatalf("StructSaver.InsertID = %q, want %q", saver.InsertID, wantInsertID)
	}
	if saver.Struct != wantRow {
		t.Fatalf("StructSaver.Struct = %#v, want %#v", saver.Struct, wantRow)
	}
}

type fakeInserter struct {
	calls int
	src   interface{}
	err   error
}

func (f *fakeInserter) Put(_ context.Context, src interface{}) error {
	f.calls++
	f.src = src
	return f.err
}
