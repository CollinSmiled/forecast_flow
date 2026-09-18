package openmeteo

const maximumForecastDays = 16

var currentForecastVariables = []string{
	"temperature_2m",
	"relative_humidity_2m",
	"apparent_temperature",
	"is_day",
	"precipitation",
	"rain",
	"showers",
	"snowfall",
	"weather_code",
	"cloud_cover",
	"pressure_msl",
	"wind_speed_10m",
	"wind_direction_10m",
	"wind_gusts_10m",
}

var operationalHourlyVariables = []string{
	"temperature_2m",
	"apparent_temperature",
	"relative_humidity_2m",
	"precipitation",
	"precipitation_probability",
	"rain",
	"showers",
	"snowfall",
	"weather_code",
	"cloud_cover",
	"pressure_msl",
	"visibility",
	"wind_speed_10m",
	"wind_direction_10m",
	"wind_gusts_10m",
	"uv_index",
	"is_day",
}

var singleRunHourlyVariables = []string{
	"temperature_2m",
	"apparent_temperature",
	"relative_humidity_2m",
	"precipitation",
	"rain",
	"showers",
	"snowfall",
	"weather_code",
	"cloud_cover",
	"pressure_msl",
	"visibility",
	"wind_speed_10m",
	"wind_direction_10m",
	"wind_gusts_10m",
	"uv_index",
	"is_day",
}

var dailyForecastVariables = []string{
	"weather_code",
	"temperature_2m_max",
	"temperature_2m_min",
	"apparent_temperature_max",
	"apparent_temperature_min",
	"precipitation_sum",
	"precipitation_probability_max",
	"precipitation_hours",
	"wind_speed_10m_max",
	"wind_gusts_10m_max",
	"wind_direction_10m_dominant",
	"sunrise",
	"sunset",
	"daylight_duration",
	"uv_index_max",
}
