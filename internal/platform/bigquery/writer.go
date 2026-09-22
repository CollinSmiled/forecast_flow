package bigquery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cloudbigquery "cloud.google.com/go/bigquery"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
)

const (
	OperationalForecastTable = "operational_forecast_events"
	ModelRunTable            = "model_run_events"
)

type inserter interface {
	Put(ctx context.Context, src interface{}) error
}

type Writer struct {
	operationalForecasts inserter
	modelRuns            inserter
}

var _ coldstore.Writer = (*Writer)(nil)

func NewWriter(client *cloudbigquery.Client, datasetID string) (*Writer, error) {
	if client == nil {
		return nil, errors.New("BigQuery client is required")
	}

	datasetID = strings.TrimSpace(datasetID)
	if datasetID == "" {
		return nil, errors.New("BigQuery dataset ID is required")
	}

	dataset := client.Dataset(datasetID)
	return newWriter(
		dataset.Table(OperationalForecastTable).Inserter(),
		dataset.Table(ModelRunTable).Inserter(),
	), nil
}

func newWriter(operationalForecasts inserter, modelRuns inserter) *Writer {
	return &Writer{
		operationalForecasts: operationalForecasts,
		modelRuns:            modelRuns,
	}
}

func (w *Writer) AppendOperationalForecast(
	ctx context.Context,
	row coldstore.OperationalForecastRow,
) error {
	if err := w.operationalForecasts.Put(ctx, &cloudbigquery.StructSaver{
		Struct:   row,
		InsertID: row.EventID,
	}); err != nil {
		return fmt.Errorf("append operational forecast to BigQuery: %w", err)
	}

	return nil
}

func (w *Writer) AppendModelRun(
	ctx context.Context,
	row coldstore.ModelRunRow,
) error {
	if err := w.modelRuns.Put(ctx, &cloudbigquery.StructSaver{
		Struct:   row,
		InsertID: row.EventID,
	}); err != nil {
		return fmt.Errorf("append model run to BigQuery: %w", err)
	}

	return nil
}
