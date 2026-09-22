package openmeteo

import (
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/verification"
)

func TestMapVerificationWeather(t *testing.T) {
	retrievedAt := time.Date(2026, time.September, 22, 10, 0, 0, 0, time.UTC)
	snapshot, err := mapVerificationWeather(3, retrievedAt, testVerificationPayload())
	if err != nil {
		t.Fatalf("map verification weather: %v", err)
	}

	if snapshot.Source != verificationWeatherSource {
		t.Errorf("source = %q, want %q", snapshot.Source, verificationWeatherSource)
	}
	if snapshot.ReferenceKind != verification.ReferenceKindReanalysis {
		t.Errorf("reference kind = %q, want reanalysis", snapshot.ReferenceKind)
	}
	if len(snapshot.Hourly) != 2 {
		t.Fatalf("hourly count = %d, want 2", len(snapshot.Hourly))
	}
	if snapshot.Hourly[0].Temperature2M == nil || *snapshot.Hourly[0].Temperature2M != 29.4 {
		t.Errorf("temperature = %v, want 29.4", snapshot.Hourly[0].Temperature2M)
	}
	if snapshot.Hourly[0].PrecipitationProbability != nil {
		t.Errorf("precipitation probability = %v, want nil", snapshot.Hourly[0].PrecipitationProbability)
	}

	expectedValidAt := time.Date(2026, time.September, 19, 17, 0, 0, 0, time.UTC)
	if !snapshot.Hourly[0].ValidAt.Equal(expectedValidAt) {
		t.Errorf("valid at = %v, want %v", snapshot.Hourly[0].ValidAt, expectedValidAt)
	}
}

func TestMapVerificationWeatherRequiresCompleteSeries(t *testing.T) {
	payload := testVerificationPayload()
	payload.Hourly.Rain = nil

	_, err := mapVerificationWeather(3, time.Now(), payload)
	if err == nil {
		t.Fatal("incomplete series error = nil, want an error")
	}
	if !strings.Contains(err.Error(), "rain") {
		t.Errorf("error = %q, want rain", err)
	}
}

func testVerificationPayload() forecastResponse {
	return forecastResponse{
		Timezone: "Asia/Jakarta",
		Hourly: forecastHourly{
			Time:                []string{"2026-09-20T00:00", "2026-09-20T01:00"},
			Temperature2M:       []*float64{float64Pointer(29.4), float64Pointer(28.9)},
			ApparentTemperature: []*float64{float64Pointer(33), float64Pointer(32.4)},
			RelativeHumidity2M:  []*float64{float64Pointer(80), float64Pointer(82)},
			Precipitation:       []*float64{float64Pointer(0), float64Pointer(0.2)},
			Rain:                []*float64{float64Pointer(0), float64Pointer(0.2)},
			Snowfall:            []*float64{float64Pointer(0), float64Pointer(0)},
			WeatherCode:         []*int{intPointer(2), intPointer(61)},
			CloudCover:          []*float64{float64Pointer(40), float64Pointer(75)},
			PressureMSL:         []*float64{float64Pointer(1009), float64Pointer(1008)},
			WindSpeed10M:        []*float64{float64Pointer(8), float64Pointer(9)},
			WindDirection10M:    []*float64{float64Pointer(220), float64Pointer(225)},
			WindGusts10M:        []*float64{float64Pointer(15), float64Pointer(18)},
			IsDay:               []*int{intPointer(0), intPointer(0)},
		},
	}
}
