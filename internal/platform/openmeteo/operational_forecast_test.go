package openmeteo

import (
	"strings"
	"testing"
)

func TestBuildOperationalForecastEndpoint(t *testing.T) {
	request := OperationalForecastRequest{
		Latitude:     -6.21462,
		Longitude:    106.84513,
		Timezone:     "Asia/Jakarta",
		ForecastDays: 10,
	}

	endpoint, err := buildOperationalForecastEndpoint(
		operationalForecastBaseURL,
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
	assertQueryValue(t, query.Get("models"), "best_match")
	assertQueryValue(t, query.Get("forecast_days"), "10")
	assertQueryValue(t, query.Get("temperature_unit"), "celsius")
	assertQueryValue(t, query.Get("wind_speed_unit"), "kmh")
	assertQueryValue(t, query.Get("precipitation_unit"), "mm")
	assertQueryValue(t, query.Get("timeformat"), "iso8601")

	assertContainsVariable(
		t,
		query.Get("current"),
		"temperature_2m",
	)
	assertContainsVariable(
		t,
		query.Get("current"),
		"weather_code",
	)
	assertContainsVariable(
		t,
		query.Get("hourly"),
		"precipitation_probability",
	)
	assertContainsVariable(
		t,
		query.Get("hourly"),
		"uv_index",
	)
	assertContainsVariable(
		t,
		query.Get("daily"),
		"sunrise",
	)
	assertContainsVariable(
		t,
		query.Get("daily"),
		"uv_index_max",
	)

	if query.Get("run") != "" {
		t.Errorf(
			"run = %q, want empty for operational forecast",
			query.Get("run"),
		)
	}
}

func TestBuildOperationalForecastEndpointRejectsInvalidRequest(
	t *testing.T,
) {
	validRequest := OperationalForecastRequest{
		Latitude:     -6.21462,
		Longitude:    106.84513,
		Timezone:     "Asia/Jakarta",
		ForecastDays: 10,
	}

	tests := []struct {
		name   string
		change func(*OperationalForecastRequest)
	}{
		{
			name: "invalid latitude",
			change: func(request *OperationalForecastRequest) {
				request.Latitude = -91
			},
		},
		{
			name: "invalid longitude",
			change: func(request *OperationalForecastRequest) {
				request.Longitude = 181
			},
		},
		{
			name: "missing timezone",
			change: func(request *OperationalForecastRequest) {
				request.Timezone = "   "
			},
		},
		{
			name: "zero forecast days",
			change: func(request *OperationalForecastRequest) {
				request.ForecastDays = 0
			},
		},
		{
			name: "too many forecast days",
			change: func(request *OperationalForecastRequest) {
				request.ForecastDays = maximumForecastDays + 1
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validRequest
			test.change(&request)

			if _, err := buildOperationalForecastEndpoint(
				operationalForecastBaseURL,
				request,
			); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestOperationalForecastVariablesRemainDistinct(t *testing.T) {
	operational := strings.Join(operationalHourlyVariables, ",")
	deterministic := strings.Join(singleRunHourlyVariables, ",")

	if !strings.Contains(
		operational,
		"precipitation_probability",
	) {
		t.Fatal("operational variables must include probability")
	}

	if strings.Contains(
		deterministic,
		"precipitation_probability",
	) {
		t.Fatal("deterministic variables must exclude probability")
	}
}
