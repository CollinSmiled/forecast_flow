package coldstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/event"
)

var ErrUnsupportedTopic = errors.New("unsupported cold-path topic")

type Processor struct {
	writer Writer
	now    func() time.Time
}

func NewProcessor(writer Writer) (*Processor, error) {
	if writer == nil {
		return nil, errors.New("cold-path writer is required")
	}

	return &Processor{
		writer: writer,
		now:    time.Now,
	}, nil
}

func (processor *Processor) Process(
	ctx context.Context,
	metadata KafkaRecordMetadata,
	payload []byte,
) error {
	return processor.ProcessBatch(ctx, []Record{{
		Metadata: metadata,
		Payload:  payload,
	}})
}

func (processor *Processor) ProcessBatch(
	ctx context.Context,
	records []Record,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ingestedAt := processor.now().UTC()
	operationalRows := make([]OperationalForecastRow, 0, len(records))
	modelRunRows := make([]ModelRunRow, 0, len(records))
	verificationRows := make([]VerificationWeatherRow, 0, len(records))

	for _, record := range records {
		switch record.Metadata.Topic {
		case event.LatestForecastTopic:
			row, err := processor.mapOperationalForecast(
				record.Metadata,
				record.Payload,
				ingestedAt,
			)
			if err != nil {
				return err
			}
			operationalRows = append(operationalRows, row)
		case event.ForecastRunTopic:
			row, err := processor.mapModelRun(
				record.Metadata,
				record.Payload,
				ingestedAt,
			)
			if err != nil {
				return err
			}
			modelRunRows = append(modelRunRows, row)
		case event.VerificationWeatherTopic:
			row, err := processor.mapVerificationWeather(
				record.Metadata,
				record.Payload,
				ingestedAt,
			)
			if err != nil {
				return err
			}
			verificationRows = append(verificationRows, row)
		default:
			return fmt.Errorf(
				"%w %q",
				ErrUnsupportedTopic,
				record.Metadata.Topic,
			)
		}
	}

	if len(operationalRows) > 0 {
		if err := processor.writer.AppendOperationalForecasts(
			ctx,
			operationalRows,
		); err != nil {
			return fmt.Errorf("append operational forecast batch: %w", err)
		}
	}

	if len(modelRunRows) > 0 {
		if err := processor.writer.AppendModelRuns(ctx, modelRunRows); err != nil {
			return fmt.Errorf("append model-run batch: %w", err)
		}
	}

	if len(verificationRows) > 0 {
		if err := processor.writer.AppendVerificationWeather(
			ctx,
			verificationRows,
		); err != nil {
			return fmt.Errorf("append verification weather batch: %w", err)
		}
	}

	return nil
}

func (processor *Processor) mapOperationalForecast(
	metadata KafkaRecordMetadata,
	payload []byte,
	ingestedAt time.Time,
) (OperationalForecastRow, error) {
	var forecastEvent event.LatestForecastEventV1
	if err := json.Unmarshal(payload, &forecastEvent); err != nil {
		return OperationalForecastRow{}, fmt.Errorf(
			"decode operational forecast event: %w",
			err,
		)
	}

	row, err := NewOperationalForecastRow(
		forecastEvent,
		metadata,
		payload,
		ingestedAt,
	)
	if err != nil {
		return OperationalForecastRow{}, fmt.Errorf(
			"map operational forecast event: %w",
			err,
		)
	}

	return row, nil
}

func (processor *Processor) mapModelRun(
	metadata KafkaRecordMetadata,
	payload []byte,
	ingestedAt time.Time,
) (ModelRunRow, error) {
	var forecastEvent event.ForecastRunEventV1
	if err := json.Unmarshal(payload, &forecastEvent); err != nil {
		return ModelRunRow{}, fmt.Errorf("decode model-run event: %w", err)
	}

	row, err := NewModelRunRow(
		forecastEvent,
		metadata,
		payload,
		ingestedAt,
	)
	if err != nil {
		return ModelRunRow{}, fmt.Errorf("map model-run event: %w", err)
	}

	return row, nil
}

func (processor *Processor) mapVerificationWeather(
	metadata KafkaRecordMetadata,
	payload []byte,
	ingestedAt time.Time,
) (VerificationWeatherRow, error) {
	var weatherEvent event.VerificationWeatherEventV1
	if err := json.Unmarshal(payload, &weatherEvent); err != nil {
		return VerificationWeatherRow{}, fmt.Errorf(
			"decode verification weather event: %w",
			err,
		)
	}

	row, err := NewVerificationWeatherRow(
		weatherEvent,
		metadata,
		payload,
		ingestedAt,
	)
	if err != nil {
		return VerificationWeatherRow{}, fmt.Errorf(
			"map verification weather event: %w",
			err,
		)
	}

	return row, nil
}
