package openmeteo

import (
	"strings"
	"testing"
	"time"

	"github.com/CollinSmiled/forecast_flow/internal/forecast"
)

func TestMapHourlyForecasts(t *testing.T) {
	run := testForecastRun(t)
	payload := testHourlyPayload()

	hourly, err := mapHourlyForecasts(run, payload)
	if err != nil {
		t.Fatalf("map hourly forecasts: %v", err)
	}

	if len(hourly) != 2 {
		t.Fatalf(
			"hourly count = %d, want 2",
			len(hourly),
		)
	}

	expectedValidAt := time.Date(
		2026,
		time.September,
		17,
		6,
		0,
		0,
		0,
		time.UTC,
	)

	if !hourly[0].ValidAt.Equal(expectedValidAt) {
		t.Errorf(
			"valid time = %v, want %v",
			hourly[0].ValidAt,
			expectedValidAt,
		)
	}

	if hourly[0].LeadTimeHours != 0 {
		t.Errorf(
			"lead time = %d, want 0",
			hourly[0].LeadTimeHours,
		)
	}

	if hourly[0].Temperature2M == nil ||
		*hourly[0].Temperature2M != 30.4 {
		t.Errorf(
			"temperature = %v, want 30.4",
			hourly[0].Temperature2M,
		)
	}

	if hourly[0].IsDay == nil || !*hourly[0].IsDay {
		t.Errorf(
			"is day = %v, want true",
			hourly[0].IsDay,
		)
	}

	if hourly[1].Temperature2M != nil {
		t.Errorf(
			"second temperature = %v, want nil",
			hourly[1].Temperature2M,
		)
	}
}

func TestMapHourlyForecastsRejectsUnequalSeries(
	t *testing.T,
) {
	run := testForecastRun(t)
	payload := testHourlyPayload()

	payload.Hourly.WindGusts10M =
		payload.Hourly.WindGusts10M[:1]

	_, err := mapHourlyForecasts(run, payload)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "wind_gusts_10m") {
		t.Errorf(
			"error = %q, want wind_gusts_10m",
			err,
		)
	}
}

func TestMapIsDayRejectsUnexpectedValue(t *testing.T) {
	value := 2

	if _, err := mapIsDay(&value); err == nil {
		t.Fatal("expected an error")
	}
}

func TestMapDailyForecasts(t *testing.T) {
	payload := testDailyPayload()

	daily, err := mapDailyForecasts(payload)
	if err != nil {
		t.Fatalf("map daily forecasts: %v", err)
	}

	if len(daily) != 1 {
		t.Fatalf(
			"daily count = %d, want 1",
			len(daily),
		)
	}

	if daily[0].Date != "2026-09-17" {
		t.Errorf(
			"date = %q, want 2026-09-17",
			daily[0].Date,
		)
	}

	if daily[0].Temperature2MMax == nil ||
		*daily[0].Temperature2MMax != 32.1 {
		t.Errorf(
			"maximum temperature = %v, want 32.1",
			daily[0].Temperature2MMax,
		)
	}

	expectedSunrise := time.Date(
		2026,
		time.September,
		16,
		22,
		45,
		0,
		0,
		time.UTC,
	)

	if daily[0].Sunrise == nil ||
		!daily[0].Sunrise.Equal(expectedSunrise) {
		t.Errorf(
			"sunrise = %v, want %v",
			daily[0].Sunrise,
			expectedSunrise,
		)
	}
}

func TestMapDailyForecastsRejectsUnequalSeries(
	t *testing.T,
) {
	payload := testDailyPayload()
	payload.Daily.UVIndexMax = nil

	_, err := mapDailyForecasts(payload)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "uv_index_max") {
		t.Errorf(
			"error = %q, want uv_index_max",
			err,
		)
	}
}

func testForecastRun(t *testing.T) forecast.Run {
	t.Helper()

	model := forecast.Model{
		ID:           "ecmwf_ifs",
		Name:         "ECMWF IFS",
		Provider:     "ECMWF",
		ResolutionKM: 9,
	}

	run, err := forecast.NewRun(
		3,
		model,
		time.Date(
			2026,
			time.September,
			17,
			6,
			0,
			0,
			0,
			time.UTC,
		),
		time.Date(
			2026,
			time.September,
			17,
			10,
			0,
			0,
			0,
			time.UTC,
		),
	)
	if err != nil {
		t.Fatalf("create forecast run: %v", err)
	}

	return run
}

func testHourlyPayload() singleRunResponse {
	return singleRunResponse{
		Timezone: "Asia/Jakarta",
		Hourly: singleRunHourly{
			Time: []string{
				"2026-09-17T13:00",
				"2026-09-17T14:00",
			},
			Temperature2M: []*float64{
				float64Pointer(30.4),
				nil,
			},
			ApparentTemperature: []*float64{
				float64Pointer(34.1),
				float64Pointer(34.5),
			},
			RelativeHumidity2M: []*float64{
				float64Pointer(79),
				float64Pointer(78),
			},
			Precipitation: []*float64{
				float64Pointer(2.1),
				float64Pointer(1.4),
			},
			PrecipitationProbability: []*float64{
				float64Pointer(75),
				float64Pointer(60),
			},
			Rain: []*float64{
				float64Pointer(2.1),
				float64Pointer(1.4),
			},
			Showers: []*float64{
				float64Pointer(0),
				float64Pointer(0),
			},
			Snowfall: []*float64{
				float64Pointer(0),
				float64Pointer(0),
			},
			WeatherCode: []*int{
				intPointer(61),
				intPointer(61),
			},
			CloudCover: []*float64{
				float64Pointer(87),
				float64Pointer(80),
			},
			PressureMSL: []*float64{
				float64Pointer(1008),
				float64Pointer(1007),
			},
			Visibility: []*float64{
				float64Pointer(12000),
				float64Pointer(11000),
			},
			WindSpeed10M: []*float64{
				float64Pointer(14.2),
				float64Pointer(15),
			},
			WindDirection10M: []*float64{
				float64Pointer(220),
				float64Pointer(225),
			},
			WindGusts10M: []*float64{
				float64Pointer(28),
				float64Pointer(30),
			},
			UVIndex: []*float64{
				float64Pointer(7.5),
				float64Pointer(6.8),
			},
			IsDay: []*int{
				intPointer(1),
				intPointer(1),
			},
		},
	}
}

func testDailyPayload() singleRunResponse {
	return singleRunResponse{
		Timezone: "Asia/Jakarta",
		Daily: singleRunDaily{
			Time: []string{
				"2026-09-17",
			},
			WeatherCode: []*int{
				intPointer(61),
			},
			Temperature2MMax: []*float64{
				float64Pointer(32.1),
			},
			Temperature2MMin: []*float64{
				float64Pointer(25.4),
			},
			ApparentTemperatureMax: []*float64{
				float64Pointer(36.2),
			},
			ApparentTemperatureMin: []*float64{
				float64Pointer(28.1),
			},
			PrecipitationSum: []*float64{
				float64Pointer(8.4),
			},
			PrecipitationProbabilityMax: []*float64{
				float64Pointer(80),
			},
			PrecipitationHours: []*float64{
				float64Pointer(4),
			},
			WindSpeed10MMax: []*float64{
				float64Pointer(18.2),
			},
			WindGusts10MMax: []*float64{
				float64Pointer(31),
			},
			WindDirection10MDominant: []*float64{
				float64Pointer(220),
			},
			Sunrise: []*string{
				stringPointer("2026-09-17T05:45"),
			},
			Sunset: []*string{
				stringPointer("2026-09-17T17:51"),
			},
			DaylightDuration: []*float64{
				float64Pointer(43560),
			},
			UVIndexMax: []*float64{
				float64Pointer(9.2),
			},
		},
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func stringPointer(value string) *string {
	return &value
}
