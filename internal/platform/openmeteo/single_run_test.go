package openmeteo

import (
	"strings"
	"testing"
	"time"
)

func TestBuildSingleRunEndpoint(t *testing.T) {
	request := SingleRunRequest{
		Latitude:  -6.21462,
		Longitude: 106.84513,
		Timezone:  "Asia/Jakarta",
		ModelID:   "ecmwf_ifs",
		ForecastRunAt: time.Date(
			2026,
			time.September,
			17,
			6,
			0,
			0,
			0,
			time.UTC,
		),
		ForecastDays: 10,
	}

	endpoint, err := buildSingleRunEndpoint(
		singleRunsBaseURL,
		request,
	)
	if err != nil {
		t.Fatalf("build endpoint: %v", err)
	}

	if endpoint.Path != "/v1/forecast" {
		t.Errorf(
			"path = %q, want %q",
			endpoint.Path,
			"/v1/forecast",
		)
	}

	query := endpoint.Query()

	assertQueryValue(t, query.Get("latitude"), "-6.21462")
	assertQueryValue(t, query.Get("longitude"), "106.84513")
	assertQueryValue(t, query.Get("timezone"), "Asia/Jakarta")
	assertQueryValue(t, query.Get("models"), "ecmwf_ifs")
	assertQueryValue(t, query.Get("run"), "2026-09-17T06:00")
	assertQueryValue(t, query.Get("forecast_days"), "10")
	assertQueryValue(t, query.Get("temperature_unit"), "celsius")
	assertQueryValue(t, query.Get("wind_speed_unit"), "kmh")
	assertQueryValue(t, query.Get("precipitation_unit"), "mm")
	assertQueryValue(t, query.Get("timeformat"), "iso8601")

	assertContainsVariable(
		t,
		query.Get("hourly"),
		"temperature_2m",
	)
	assertContainsVariable(
		t,
		query.Get("hourly"),
		"uv_index",
	)

	if strings.Contains(
		query.Get("hourly"),
		"precipitation_probability",
	) {
		t.Error(
			"deterministic single-run request must not include precipitation_probability",
		)
	}

	if query.Get("daily") != "" {
		t.Errorf(
			"daily = %q, want empty for local-time single run",
			query.Get("daily"),
		)
	}
}

func TestBuildSingleRunEndpointRejectsInvalidRequest(
	t *testing.T,
) {
	validRequest := SingleRunRequest{
		Latitude:  -6.21462,
		Longitude: 106.84513,
		Timezone:  "Asia/Jakarta",
		ModelID:   "ecmwf_ifs",
		ForecastRunAt: time.Date(
			2026,
			time.September,
			17,
			6,
			0,
			0,
			0,
			time.UTC,
		),
		ForecastDays: 10,
	}

	tests := []struct {
		name   string
		change func(*SingleRunRequest)
	}{
		{
			name: "invalid latitude",
			change: func(request *SingleRunRequest) {
				request.Latitude = 91
			},
		},
		{
			name: "missing timezone",
			change: func(request *SingleRunRequest) {
				request.Timezone = ""
			},
		},
		{
			name: "missing model",
			change: func(request *SingleRunRequest) {
				request.ModelID = ""
			},
		},
		{
			name: "partial-hour run",
			change: func(request *SingleRunRequest) {
				request.ForecastRunAt = request.
					ForecastRunAt.
					Add(30 * time.Minute)
			},
		},
		{
			name: "too many forecast days",
			change: func(request *SingleRunRequest) {
				request.ForecastDays = 17
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validRequest
			test.change(&request)

			if _, err := buildSingleRunEndpoint(
				singleRunsBaseURL,
				request,
			); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func assertQueryValue(
	t *testing.T,
	got string,
	want string,
) {
	t.Helper()

	if got != want {
		t.Errorf("query value = %q, want %q", got, want)
	}
}

func assertContainsVariable(
	t *testing.T,
	value string,
	want string,
) {
	t.Helper()

	for _, variable := range strings.Split(value, ",") {
		if variable == want {
			return
		}
	}

	t.Errorf(
		"variables %q do not contain %q",
		value,
		want,
	)
}
