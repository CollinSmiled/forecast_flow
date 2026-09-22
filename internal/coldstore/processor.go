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
	if err := ctx.Err(); err != nil {
		return err
	}

	switch metadata.Topic {
	case event.LatestForecastTopic:
		return processor.processOperationalForecast(
			ctx,
			metadata,
			payload,
		)
	case event.ForecastRunTopic:
		return processor.processModelRun(ctx, metadata, payload)
	default:
		return fmt.Errorf("%w %q", ErrUnsupportedTopic, metadata.Topic)
	}
}

func (processor *Processor) processOperationalForecast(
	ctx context.Context,
	metadata KafkaRecordMetadata,
	payload []byte,
) error {
	var forecastEvent event.LatestForecastEventV1
	if err := json.Unmarshal(payload, &forecastEvent); err != nil {
		return fmt.Errorf("decode operational forecast event: %w", err)
	}

	row, err := NewOperationalForecastRow(
		forecastEvent,
		metadata,
		payload,
		processor.now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("map operational forecast event: %w", err)
	}

	if err := processor.writer.AppendOperationalForecast(ctx, row); err != nil {
		return fmt.Errorf("append operational forecast event: %w", err)
	}

	return nil
}

func (processor *Processor) processModelRun(
	ctx context.Context,
	metadata KafkaRecordMetadata,
	payload []byte,
) error {
	var forecastEvent event.ForecastRunEventV1
	if err := json.Unmarshal(payload, &forecastEvent); err != nil {
		return fmt.Errorf("decode model-run event: %w", err)
	}

	row, err := NewModelRunRow(
		forecastEvent,
		metadata,
		payload,
		processor.now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("map model-run event: %w", err)
	}

	if err := processor.writer.AppendModelRun(ctx, row); err != nil {
		return fmt.Errorf("append model-run event: %w", err)
	}

	return nil
}
