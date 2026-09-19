package openmeteo

type forecastResponse struct {
	Latitude             float64         `json:"latitude"`
	Longitude            float64         `json:"longitude"`
	GenerationTimeMS     float64         `json:"generationtime_ms"`
	UTCOffsetSeconds     int             `json:"utc_offset_seconds"`
	Timezone             string          `json:"timezone"`
	TimezoneAbbreviation string          `json:"timezone_abbreviation"`
	Elevation            *float64        `json:"elevation"`
	Current              forecastCurrent `json:"current"`
	Hourly               forecastHourly  `json:"hourly"`
	Daily                forecastDaily   `json:"daily"`
}

type forecastCurrent struct {
	Time                string   `json:"time"`
	IntervalSeconds     int      `json:"interval"`
	Temperature2M       *float64 `json:"temperature_2m"`
	RelativeHumidity2M  *float64 `json:"relative_humidity_2m"`
	ApparentTemperature *float64 `json:"apparent_temperature"`
	IsDay               *int     `json:"is_day"`
	Precipitation       *float64 `json:"precipitation"`
	Rain                *float64 `json:"rain"`
	Showers             *float64 `json:"showers"`
	Snowfall            *float64 `json:"snowfall"`
	WeatherCode         *int     `json:"weather_code"`
	CloudCover          *float64 `json:"cloud_cover"`
	PressureMSL         *float64 `json:"pressure_msl"`
	WindSpeed10M        *float64 `json:"wind_speed_10m"`
	WindDirection10M    *float64 `json:"wind_direction_10m"`
	WindGusts10M        *float64 `json:"wind_gusts_10m"`
}

type forecastHourly struct {
	Time                     []string   `json:"time"`
	Temperature2M            []*float64 `json:"temperature_2m"`
	ApparentTemperature      []*float64 `json:"apparent_temperature"`
	RelativeHumidity2M       []*float64 `json:"relative_humidity_2m"`
	Precipitation            []*float64 `json:"precipitation"`
	PrecipitationProbability []*float64 `json:"precipitation_probability"`
	Rain                     []*float64 `json:"rain"`
	Showers                  []*float64 `json:"showers"`
	Snowfall                 []*float64 `json:"snowfall"`
	WeatherCode              []*int     `json:"weather_code"`
	CloudCover               []*float64 `json:"cloud_cover"`
	PressureMSL              []*float64 `json:"pressure_msl"`
	Visibility               []*float64 `json:"visibility"`
	WindSpeed10M             []*float64 `json:"wind_speed_10m"`
	WindDirection10M         []*float64 `json:"wind_direction_10m"`
	WindGusts10M             []*float64 `json:"wind_gusts_10m"`
	UVIndex                  []*float64 `json:"uv_index"`
	IsDay                    []*int     `json:"is_day"`
}

type forecastDaily struct {
	Time                        []string   `json:"time"`
	WeatherCode                 []*int     `json:"weather_code"`
	Temperature2MMax            []*float64 `json:"temperature_2m_max"`
	Temperature2MMin            []*float64 `json:"temperature_2m_min"`
	ApparentTemperatureMax      []*float64 `json:"apparent_temperature_max"`
	ApparentTemperatureMin      []*float64 `json:"apparent_temperature_min"`
	PrecipitationSum            []*float64 `json:"precipitation_sum"`
	PrecipitationProbabilityMax []*float64 `json:"precipitation_probability_max"`
	PrecipitationHours          []*float64 `json:"precipitation_hours"`
	WindSpeed10MMax             []*float64 `json:"wind_speed_10m_max"`
	WindGusts10MMax             []*float64 `json:"wind_gusts_10m_max"`
	WindDirection10MDominant    []*float64 `json:"wind_direction_10m_dominant"`
	Sunrise                     []*string  `json:"sunrise"`
	Sunset                      []*string  `json:"sunset"`
	DaylightDuration            []*float64 `json:"daylight_duration"`
	UVIndexMax                  []*float64 `json:"uv_index_max"`
}
