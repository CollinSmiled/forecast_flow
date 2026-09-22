package verification

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
)

const ReferenceKindReanalysis = "reanalysis"

type HourlyWeather struct {
	ValidAt time.Time
	forecast.WeatherMetrics
}

type Snapshot struct {
	LocationID    int64
	Source        string
	ReferenceKind string
	RetrievedAt   time.Time
	Timezone      string
	Hourly        []HourlyWeather
}

func (snapshot Snapshot) Validate() error {
	if snapshot.LocationID < 1 {
		return errors.New("location ID must be greater than zero")
	}
	if strings.TrimSpace(snapshot.Source) == "" {
		return errors.New("verification source is required")
	}
	if snapshot.ReferenceKind != ReferenceKindReanalysis {
		return fmt.Errorf(
			"unsupported verification reference kind %q",
			snapshot.ReferenceKind,
		)
	}
	if snapshot.RetrievedAt.IsZero() {
		return errors.New("retrieval time is required")
	}
	if strings.TrimSpace(snapshot.Timezone) == "" {
		return errors.New("timezone is required")
	}
	if len(snapshot.Hourly) == 0 {
		return errors.New("hourly verification weather is required")
	}

	for index, item := range snapshot.Hourly {
		if item.ValidAt.IsZero() {
			return fmt.Errorf("hourly record at index %d requires a valid time", index)
		}
		if index > 0 && !item.ValidAt.After(snapshot.Hourly[index-1].ValidAt) {
			return errors.New("hourly verification times must be strictly increasing")
		}
	}

	return nil
}
