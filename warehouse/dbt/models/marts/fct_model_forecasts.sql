select
    event_id,
    location_id,
    model_id,
    model_name,
    provider,
    resolution_km,
    forecast_run_at,
    retrieved_at,
    valid_at,
    date(valid_at) as forecast_date_utc,
    lead_time_hours,
    case
        when lead_time_hours <= 6 then '00-06h'
        when lead_time_hours <= 24 then '07-24h'
        when lead_time_hours <= 72 then '25-72h'
        else '73h+'
    end as lead_time_bucket,
    temperature_2m,
    apparent_temperature,
    relative_humidity_2m,
    precipitation,
    precipitation_probability,
    rain,
    showers,
    snowfall,
    weather_code,
    cloud_cover,
    pressure_msl,
    visibility,
    wind_speed_10m,
    wind_direction_10m,
    wind_gusts_10m,
    uv_index,
    is_day,
    ingested_at
from {{ ref('int_model_run_hourly_forecasts') }}
