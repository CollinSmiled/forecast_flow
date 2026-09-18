package forecast

import (
	"fmt"
	"strings"
	"time"
)

type Model struct {
	ID           string
	Name         string
	Provider     string
	ResolutionKM float64
}

func (model Model) Validate() error {
	if strings.TrimSpace(model.ID) == "" {
		return fmt.Errorf("model ID is required")
	}

	if strings.TrimSpace(model.Name) == "" {
		return fmt.Errorf("model name is required")
	}

	if strings.TrimSpace(model.Provider) == "" {
		return fmt.Errorf("model provider is required")
	}

	if model.ResolutionKM <= 0 {
		return fmt.Errorf("model resolution must be greater than zero")
	}

	return nil
}

type Run struct {
	LocationID    int64
	Model         Model
	ForecastRunAt time.Time
	RetrievedAt   time.Time
}

func NewRun(
	locationID int64,
	model Model,
	forecastRunAt time.Time,
	retrievedAt time.Time,
) (Run, error) {
	if locationID <= 0 {
		return Run{}, fmt.Errorf("location ID must be greater than zero")
	}

	if err := model.Validate(); err != nil {
		return Run{}, fmt.Errorf("validate model: %w", err)
	}

	if forecastRunAt.IsZero() {
		return Run{}, fmt.Errorf("forecast run time is required")
	}

	if retrievedAt.IsZero() {
		return Run{}, fmt.Errorf("retrieval time is required")
	}

	forecastRunAt = forecastRunAt.UTC()
	retrievedAt = retrievedAt.UTC()

	if retrievedAt.Before(forecastRunAt) {
		return Run{}, fmt.Errorf(
			"retrieval time cannot be before forecast run time",
		)
	}

	return Run{
		LocationID:    locationID,
		Model:         model,
		ForecastRunAt: forecastRunAt,
		RetrievedAt:   retrievedAt,
	}, nil
}

func (run Run) LeadTimeHours(validAt time.Time) (int, error) {
	if validAt.IsZero() {
		return 0, fmt.Errorf("valid time is required")
	}

	leadTime := validAt.UTC().Sub(run.ForecastRunAt)

	if leadTime < 0 {
		return 0, fmt.Errorf("valid time cannot be before forecast run time")
	}

	if leadTime%time.Hour != 0 {
		return 0, fmt.Errorf("valid time must align to a whole hour")
	}

	return int(leadTime / time.Hour), nil
}
