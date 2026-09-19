CREATE TABLE public.latest_operational_forecasts (
    location_id BIGINT PRIMARY KEY
        REFERENCES public.locations (location_id),

    event_id TEXT NOT NULL UNIQUE,
    schema_version INTEGER NOT NULL,
    source TEXT NOT NULL,
    retrieved_at TIMESTAMPTZ NOT NULL,
    timezone TEXT NOT NULL,

    current_valid_at TIMESTAMPTZ NOT NULL,
    current_interval_seconds INTEGER NOT NULL,

    current_temperature_2m DOUBLE PRECISION,
    current_apparent_temperature DOUBLE PRECISION,
    current_relative_humidity_2m DOUBLE PRECISION,

    current_precipitation DOUBLE PRECISION,
    current_precipitation_probability DOUBLE PRECISION,
    current_rain DOUBLE PRECISION,
    current_showers DOUBLE PRECISION,
    current_snowfall DOUBLE PRECISION,

    current_weather_code INTEGER,
    current_cloud_cover DOUBLE PRECISION,
    current_pressure_msl DOUBLE PRECISION,
    current_visibility DOUBLE PRECISION,

    current_wind_speed_10m DOUBLE PRECISION,
    current_wind_direction_10m DOUBLE PRECISION,
    current_wind_gusts_10m DOUBLE PRECISION,

    current_uv_index DOUBLE PRECISION,
    current_is_day BOOLEAN,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_latest_operational_schema_version
        CHECK (schema_version > 0),

    CONSTRAINT chk_latest_operational_source
        CHECK (source <> ''),

    CONSTRAINT chk_latest_operational_timezone
        CHECK (timezone <> ''),

    CONSTRAINT chk_latest_operational_current_interval
        CHECK (current_interval_seconds > 0)
);

CREATE TABLE public.latest_operational_hourly (
    location_id BIGINT NOT NULL
        REFERENCES public.latest_operational_forecasts (location_id)
        ON DELETE CASCADE,

    valid_at TIMESTAMPTZ NOT NULL,

    temperature_2m DOUBLE PRECISION,
    apparent_temperature DOUBLE PRECISION,
    relative_humidity_2m DOUBLE PRECISION,

    precipitation DOUBLE PRECISION,
    precipitation_probability DOUBLE PRECISION,
    rain DOUBLE PRECISION,
    showers DOUBLE PRECISION,
    snowfall DOUBLE PRECISION,

    weather_code INTEGER,
    cloud_cover DOUBLE PRECISION,
    pressure_msl DOUBLE PRECISION,
    visibility DOUBLE PRECISION,

    wind_speed_10m DOUBLE PRECISION,
    wind_direction_10m DOUBLE PRECISION,
    wind_gusts_10m DOUBLE PRECISION,

    uv_index DOUBLE PRECISION,
    is_day BOOLEAN,

    PRIMARY KEY (location_id, valid_at)
);

CREATE TABLE public.latest_operational_daily (
    location_id BIGINT NOT NULL
        REFERENCES public.latest_operational_forecasts (location_id)
        ON DELETE CASCADE,

    forecast_date DATE NOT NULL,

    weather_code INTEGER,

    temperature_2m_max DOUBLE PRECISION,
    temperature_2m_min DOUBLE PRECISION,
    apparent_temperature_max DOUBLE PRECISION,
    apparent_temperature_min DOUBLE PRECISION,

    precipitation_sum DOUBLE PRECISION,
    precipitation_probability_max DOUBLE PRECISION,
    precipitation_hours DOUBLE PRECISION,

    wind_speed_10m_max DOUBLE PRECISION,
    wind_gusts_10m_max DOUBLE PRECISION,
    wind_direction_10m_dominant DOUBLE PRECISION,

    sunrise TIMESTAMPTZ,
    sunset TIMESTAMPTZ,
    daylight_duration_seconds DOUBLE PRECISION,
    uv_index_max DOUBLE PRECISION,

    PRIMARY KEY (location_id, forecast_date)
);
