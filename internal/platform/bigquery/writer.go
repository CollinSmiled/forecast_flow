package bigquery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	cloudbigquery "cloud.google.com/go/bigquery"
	"google.golang.org/api/googleapi"

	"github.com/CollinSmiled/forecast_flow/internal/coldstore"
	"github.com/CollinSmiled/forecast_flow/internal/referencedata"
)

const (
	OperationalForecastTable = "operational_forecast_events"
	ModelRunTable            = "model_run_events"
	LocationsTable           = "locations"
)

type loader interface {
	Load(ctx context.Context, jobID string, data []byte) error
}

type tableLoader struct {
	client           *cloudbigquery.Client
	table            *cloudbigquery.Table
	location         string
	writeDisposition cloudbigquery.TableWriteDisposition
}

type Writer struct {
	operationalForecasts loader
	modelRuns            loader
}

type ReferenceWriter struct {
	locations loader
}

var _ coldstore.Writer = (*Writer)(nil)
var _ referencedata.LocationWriter = (*ReferenceWriter)(nil)

func NewWriter(
	client *cloudbigquery.Client,
	datasetID string,
	location string,
) (*Writer, error) {
	if client == nil {
		return nil, errors.New("BigQuery client is required")
	}

	datasetID = strings.TrimSpace(datasetID)
	if datasetID == "" {
		return nil, errors.New("BigQuery dataset ID is required")
	}

	location = strings.TrimSpace(location)
	if location == "" {
		return nil, errors.New("BigQuery location is required")
	}

	dataset := client.Dataset(datasetID)
	return newWriter(
		&tableLoader{
			client:           client,
			table:            dataset.Table(OperationalForecastTable),
			location:         location,
			writeDisposition: cloudbigquery.WriteAppend,
		},
		&tableLoader{
			client:           client,
			table:            dataset.Table(ModelRunTable),
			location:         location,
			writeDisposition: cloudbigquery.WriteAppend,
		},
	), nil
}

func NewReferenceWriter(
	client *cloudbigquery.Client,
	datasetID string,
	location string,
) (*ReferenceWriter, error) {
	if client == nil {
		return nil, errors.New("BigQuery client is required")
	}

	datasetID = strings.TrimSpace(datasetID)
	if datasetID == "" {
		return nil, errors.New("BigQuery dataset ID is required")
	}

	location = strings.TrimSpace(location)
	if location == "" {
		return nil, errors.New("BigQuery location is required")
	}

	return newReferenceWriter(&tableLoader{
		client:           client,
		table:            client.Dataset(datasetID).Table(LocationsTable),
		location:         location,
		writeDisposition: cloudbigquery.WriteTruncate,
	}), nil
}

func newWriter(operationalForecasts loader, modelRuns loader) *Writer {
	return &Writer{
		operationalForecasts: operationalForecasts,
		modelRuns:            modelRuns,
	}
}

func newReferenceWriter(locations loader) *ReferenceWriter {
	return &ReferenceWriter{locations: locations}
}

func (w *Writer) AppendOperationalForecasts(
	ctx context.Context,
	rows []coldstore.OperationalForecastRow,
) error {
	if err := loadRows(
		ctx,
		w.operationalForecasts,
		"operational",
		rows,
	); err != nil {
		return fmt.Errorf("append operational forecast batch to BigQuery: %w", err)
	}

	return nil
}

func (w *Writer) AppendModelRuns(
	ctx context.Context,
	rows []coldstore.ModelRunRow,
) error {
	if err := loadRows(ctx, w.modelRuns, "model_run", rows); err != nil {
		return fmt.Errorf("append model-run batch to BigQuery: %w", err)
	}

	return nil
}

func (w *ReferenceWriter) ReplaceLocations(
	ctx context.Context,
	rows []referencedata.LocationRow,
) error {
	if len(rows) == 0 {
		return errors.New("replace location snapshot: rows are required")
	}

	if err := loadRows(ctx, w.locations, "locations", rows); err != nil {
		return fmt.Errorf("replace location snapshot in BigQuery: %w", err)
	}

	return nil
}

func loadRows[T any](
	ctx context.Context,
	destination loader,
	jobPrefix string,
	rows []T,
) error {
	if len(rows) == 0 {
		return nil
	}

	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	for index := range rows {
		if err := encoder.Encode(rows[index]); err != nil {
			return fmt.Errorf("encode row %d: %w", index, err)
		}
	}

	payload := data.Bytes()
	if err := destination.Load(
		ctx,
		batchJobID(jobPrefix, payload),
		payload,
	); err != nil {
		return err
	}

	return nil
}

func batchJobID(prefix string, data []byte) string {
	digest := sha256.Sum256(data)
	return "forecast_flow_" + prefix + "_" + hex.EncodeToString(digest[:16])
}

func (destination *tableLoader) Load(
	ctx context.Context,
	jobID string,
	data []byte,
) error {
	source := cloudbigquery.NewReaderSource(bytes.NewReader(data))
	source.SourceFormat = cloudbigquery.JSON

	load := destination.table.LoaderFrom(source)
	load.JobID = jobID
	load.Location = destination.location
	load.CreateDisposition = cloudbigquery.CreateNever
	load.WriteDisposition = destination.writeDisposition

	job, err := load.Run(ctx)
	if err != nil {
		var apiError *googleapi.Error
		if !errors.As(err, &apiError) || apiError.Code != 409 {
			return fmt.Errorf("start BigQuery load job %q: %w", jobID, err)
		}

		job, err = destination.client.JobFromIDLocation(
			ctx,
			jobID,
			destination.location,
		)
		if err != nil {
			return fmt.Errorf("find existing BigQuery load job %q: %w", jobID, err)
		}
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for BigQuery load job %q: %w", jobID, err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("BigQuery load job %q failed: %w", jobID, err)
	}

	return nil
}
