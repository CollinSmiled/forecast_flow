package openmeteo

import (
	"strings"
	"testing"
	"time"
)

func TestMapOperationalForecast(t *testing.T) {
	payload := testOperationalPayload()
	retrievedAt := time.Date(
		2026,
		time.September,
		17,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	snapshot, err := mapOperationalForecast(
		3,
		retrievedAt,
		payload,
	)
	if err != nil {
		t.Fatalf("map operational forecast: %v", err)
	}

	if snapshot.LocationID != 3 {
		t.Errorf(
			"location ID = %d, want 3",
			snapshot.LocationID,
		)
	}

	if snapshot.Source != operationalForecastSource {
		t.Errorf(
			"source = %q, want %q",
			snapshot.Source,
			operationalForecastSource,
		)
	}

	expectedCurrentTime := time.Date(
		2026,
		time.September,
		17,
		6,
		15,
		0,
		0,
		time.UTC,
	)

	if !snapshot.Current.ValidAt.Equal(expectedCurrentTime) {
		t.Errorf(
			"current valid time = %v, want %v",
			snapshot.Current.ValidAt,
			expectedCurrentTime,
		)
	}

	if snapshot.Current.Temperature2M == nil ||
		*snapshot.Current.Temperature2M != 30.8 {
		t.Errorf(
			"current temperature = %v, want 30.8",
			snapshot.Current.Temperature2M,
		)
	}

	if len(snapshot.Hourly) != 2 {
		t.Fatalf(
			"hourly count = %d, want 2",
			len(snapshot.Hourly),
		)
	}

	if snapshot.Hourly[0].PrecipitationProbability == nil ||
		*snapshot.Hourly[0].PrecipitationProbability != 75 {
		t.Errorf(
			"probability = %v, want 75",
			snapshot.Hourly[0].PrecipitationProbability,
		)
	}

	if len(snapshot.Daily) != 1 {
		t.Fatalf(
			"daily count = %d, want 1",
			len(snapshot.Daily),
		)
	}

	if snapshot.Daily[0].Date != "2026-09-17" {
		t.Errorf(
			"daily date = %q, want 2026-09-17",
			snapshot.Daily[0].Date,
		)
	}
}

func TestMapOperationalForecastRequiresProbabilitySeries(
	t *testing.T,
) {
	payload := testOperationalPayload()
	payload.Hourly.PrecipitationProbability = nil

	_, err := mapOperationalForecast(
		3,
		time.Now(),
		payload,
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(
		err.Error(),
		"precipitation_probability",
	) {
		t.Errorf(
			"error = %q, want precipitation_probability",
			err,
		)
	}
}

func testOperationalPayload() forecastResponse {
	payload := testHourlyPayload()
	payload.Daily = testDailyPayload().Daily
	payload.Current = forecastCurrent{
		Time:                "2026-09-17T13:15",
		IntervalSeconds:     900,
		Temperature2M:       float64Pointer(30.8),
		RelativeHumidity2M:  float64Pointer(77),
		ApparentTemperature: float64Pointer(34.2),
		IsDay:               intPointer(1),
		Precipitation:       float64Pointer(0.2),
		Rain:                float64Pointer(0.2),
		Showers:             float64Pointer(0),
		Snowfall:            float64Pointer(0),
		WeatherCode:         intPointer(2),
		CloudCover:          float64Pointer(45),
		PressureMSL:         float64Pointer(1009),
		WindSpeed10M:        float64Pointer(12.4),
		WindDirection10M:    float64Pointer(210),
		WindGusts10M:        float64Pointer(24),
	}

	return payload
}
