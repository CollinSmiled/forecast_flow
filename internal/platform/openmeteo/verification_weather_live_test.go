package openmeteo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/location"
	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

func TestVerificationWeatherClientLive(t *testing.T) {
	if os.Getenv("OPENMETEO_LIVE_TEST") != "1" {
		t.Skip("set OPENMETEO_LIVE_TEST=1 to call Open-Meteo")
	}

	client := NewVerificationWeatherClient()
	date := time.Now().UTC().AddDate(0, 0, -7)
	snapshot, err := client.FetchVerificationWeather(
		context.Background(),
		location.Location{
			ID:        1,
			City:      "Jakarta",
			Country:   "Indonesia",
			Latitude:  -6.21462,
			Longitude: 106.84513,
			Timezone:  "Asia/Jakarta",
		},
		date,
		date,
	)
	if err != nil {
		t.Fatalf("fetch live verification weather: %v", err)
	}

	if snapshot.ReferenceKind != verification.ReferenceKindReanalysis {
		t.Errorf("reference kind = %q, want reanalysis", snapshot.ReferenceKind)
	}
	if len(snapshot.Hourly) != 24 {
		t.Errorf("hourly records = %d, want 24", len(snapshot.Hourly))
	}
}
