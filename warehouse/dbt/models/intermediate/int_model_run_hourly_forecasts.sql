with expanded as (
    select
        events.event_id,
        events.location_id,
        events.model_id,
        events.model_name,
        events.provider,
        events.resolution_km,
        events.forecast_run_at,
        events.retrieved_at,
        events.event_occurred_at,
        events.ingested_at,
        safe_cast(json_value(hourly, '$.valid_at') as timestamp) as valid_at,
        safe_cast(json_value(hourly, '$.lead_time_hours') as int64) as lead_time_hours,
        hourly
    from {{ ref('stg_model_run_events') }} as events
    cross join unnest(json_query_array(events.hourly_forecasts_json)) as hourly
)

select
    event_id,
    location_id,
    model_id,
    model_name,
    provider,
    resolution_km,
    forecast_run_at,
    retrieved_at,
    event_occurred_at,
    ingested_at,
    valid_at,
    lead_time_hours,
    safe_cast(json_value(hourly, '$.temperature_2m') as float64) as temperature_2m,
    safe_cast(json_value(hourly, '$.apparent_temperature') as float64) as apparent_temperature,
    safe_cast(json_value(hourly, '$.relative_humidity_2m') as float64) as relative_humidity_2m,
    safe_cast(json_value(hourly, '$.precipitation') as float64) as precipitation,
    safe_cast(json_value(hourly, '$.precipitation_probability') as float64) as precipitation_probability,
    safe_cast(json_value(hourly, '$.rain') as float64) as rain,
    safe_cast(json_value(hourly, '$.showers') as float64) as showers,
    safe_cast(json_value(hourly, '$.snowfall') as float64) as snowfall,
    safe_cast(json_value(hourly, '$.weather_code') as int64) as weather_code,
    safe_cast(json_value(hourly, '$.cloud_cover') as float64) as cloud_cover,
    safe_cast(json_value(hourly, '$.pressure_msl') as float64) as pressure_msl,
    safe_cast(json_value(hourly, '$.visibility') as float64) as visibility,
    safe_cast(json_value(hourly, '$.wind_speed_10m') as float64) as wind_speed_10m,
    safe_cast(json_value(hourly, '$.wind_direction_10m') as float64) as wind_direction_10m,
    safe_cast(json_value(hourly, '$.wind_gusts_10m') as float64) as wind_gusts_10m,
    safe_cast(json_value(hourly, '$.uv_index') as float64) as uv_index,
    safe_cast(json_value(hourly, '$.is_day') as bool) as is_day
from expanded
