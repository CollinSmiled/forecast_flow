package openmeteo

import (
	"context"
	"os"
	"testing"

	"github.com/CollinSmiled/forecast_flow/internal/location"
)

func TestOperationalForecastClientLive(t *testing.T) {
	if os.Getenv("OPENMETEO_LIVE_TEST") != "1" {
		t.Skip("set OPENMETEO_LIVE_TEST=1 to call Open-Meteo")
	}

	client := NewOperationalForecastClient()

	snapshot, err := client.FetchOperationalForecast(
		context.Background(),
		location.Location{
			ID:        1,
			City:      "Jakarta",
			Country:   "Indonesia",
			Latitude:  -6.21462,
			Longitude: 106.84513,
			Timezone:  "Asia/Jakarta",
		},
		10,
	)
	if err != nil {
		t.Fatalf("fetch live operational forecast: %v", err)
	}

	if snapshot.Current.ValidAt.IsZero() {
		t.Error("current valid time is missing")
	}

	if len(snapshot.Hourly) == 0 {
		t.Error("hourly forecasts are missing")
	}

	if len(snapshot.Daily) == 0 {
		t.Error("daily forecasts are missing")
	}
}
