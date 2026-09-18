package forecast

import (
	"fmt"
	"strings"
	"time"
)

// WeatherMetrics contains weather measurements shared by current and
// hourly forecast records.
//
// Units used by the application contract:
//   - temperatures: degrees Celsius
//   - humidity, precipitation probability, cloud cover: percent
//   - precipitation, rain, showers, snowfall: millimetres
//   - pressure: hectopascals
//   - visibility: metres
//   - wind speed and gusts: kilometres per hour
//   - wind direction: degrees
type WeatherMetrics struct {
	Temperature2M            *float64
	ApparentTemperature      *float64
	RelativeHumidity2M       *float64
	Precipitation            *float64
	PrecipitationProbability *float64
	Rain                     *float64
	Showers                  *float64
	Snowfall                 *float64
	WeatherCode              *int
	CloudCover               *float64
	PressureMSL              *float64
	Visibility               *float64
	WindSpeed10M             *float64
	WindDirection10M         *float64
	WindGusts10M             *float64
	UVIndex                  *float64
	IsDay                    *bool
}

// HourlyForecast represents one predicted valid hour from an exact
// model run.
type HourlyForecast struct {
	ValidAt       time.Time
	LeadTimeHours int
	WeatherMetrics
}

// OperationalHourlyForecast represents one hour from the latest
// operational forecast. It has no exact model-run identity.
type OperationalHourlyForecast struct {
	ValidAt time.Time
	WeatherMetrics
}

// CurrentConditions represents the provider's current conditions at
// one valid instant.
type CurrentConditions struct {
	ValidAt         time.Time
	IntervalSeconds int
	WeatherMetrics
}

func NewHourlyForecast(
	run Run,
	validAt time.Time,
) (HourlyForecast, error) {
	leadTimeHours, err := run.LeadTimeHours(validAt)
	if err != nil {
		return HourlyForecast{}, fmt.Errorf(
			"calculate hourly forecast lead time: %w",
			err,
		)
	}

	return HourlyForecast{
		ValidAt:       validAt.UTC(),
		LeadTimeHours: leadTimeHours,
	}, nil
}

// DailyForecast represents one predicted calendar day in the location's
// local timezone. Date uses the ISO 8601 YYYY-MM-DD format.
type DailyForecast struct {
	Date string

	WeatherCode                 *int
	Temperature2MMax            *float64
	Temperature2MMin            *float64
	ApparentTemperatureMax      *float64
	ApparentTemperatureMin      *float64
	PrecipitationSum            *float64
	PrecipitationProbabilityMax *float64
	PrecipitationHours          *float64
	WindSpeed10MMax             *float64
	WindGusts10MMax             *float64
	WindDirection10MDominant    *float64
	Sunrise                     *time.Time
	Sunset                      *time.Time
	DaylightDurationSeconds     *float64
	UVIndexMax                  *float64
}

func NewDailyForecast(date string) (DailyForecast, error) {
	date = strings.TrimSpace(date)

	if date == "" {
		return DailyForecast{}, fmt.Errorf("forecast date is required")
	}

	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return DailyForecast{}, fmt.Errorf(
			"forecast date must use YYYY-MM-DD format: %w",
			err,
		)
	}

	return DailyForecast{
		Date: date,
	}, nil
}

// ForecastRun represents the exact hourly forecast produced by one
// deterministic model initialization for one location.
type ForecastRun struct {
	Run    Run
	Hourly []HourlyForecast
}

// OperationalForecastSnapshot is the complete forecast visible to the
// application at one retrieval time.
type OperationalForecastSnapshot struct {
	LocationID  int64
	Source      string
	RetrievedAt time.Time
	Timezone    string
	Current     CurrentConditions
	Hourly      []OperationalHourlyForecast
	Daily       []DailyForecast
}
