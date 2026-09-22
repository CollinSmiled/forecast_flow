with matched as (
    select
        forecasts.event_id as forecast_event_id,
        verification.event_id as verification_event_id,
        forecasts.location_id,
        forecasts.model_id,
        forecasts.model_name,
        forecasts.provider,
        forecasts.resolution_km,
        forecasts.forecast_run_at,
        forecasts.retrieved_at as forecast_retrieved_at,
        forecasts.valid_at,
        forecasts.forecast_date_utc,
        forecasts.lead_time_hours,
        forecasts.lead_time_bucket,
        verification.source as verification_source,
        verification.reference_kind,
        verification.retrieved_at as verification_retrieved_at,
        verification.verification_version_count,
        forecasts.temperature_2m as forecast_temperature_2m,
        verification.temperature_2m as reference_temperature_2m,
        forecasts.apparent_temperature as forecast_apparent_temperature,
        verification.apparent_temperature as reference_apparent_temperature,
        forecasts.relative_humidity_2m as forecast_relative_humidity_2m,
        verification.relative_humidity_2m as reference_relative_humidity_2m,
        forecasts.precipitation as forecast_precipitation,
        verification.precipitation as reference_precipitation,
        forecasts.precipitation_probability as forecast_precipitation_probability,
        forecasts.rain as forecast_rain,
        verification.rain as reference_rain,
        forecasts.snowfall as forecast_snowfall,
        verification.snowfall as reference_snowfall,
        forecasts.weather_code as forecast_weather_code,
        verification.weather_code as reference_weather_code,
        forecasts.cloud_cover as forecast_cloud_cover,
        verification.cloud_cover as reference_cloud_cover,
        forecasts.pressure_msl as forecast_pressure_msl,
        verification.pressure_msl as reference_pressure_msl,
        forecasts.visibility as forecast_visibility,
        verification.visibility as reference_visibility,
        forecasts.wind_speed_10m as forecast_wind_speed_10m,
        verification.wind_speed_10m as reference_wind_speed_10m,
        forecasts.wind_direction_10m as forecast_wind_direction_10m,
        verification.wind_direction_10m as reference_wind_direction_10m,
        forecasts.wind_gusts_10m as forecast_wind_gusts_10m,
        verification.wind_gusts_10m as reference_wind_gusts_10m,
        forecasts.uv_index as forecast_uv_index,
        verification.uv_index as reference_uv_index,
        forecasts.is_day as forecast_is_day,
        verification.is_day as reference_is_day,
        forecasts.ingested_at as forecast_ingested_at,
        verification.ingested_at as verification_ingested_at
    from {{ ref('fct_model_forecasts') }} as forecasts
    inner join {{ ref('fct_verification_weather') }} as verification
        on forecasts.location_id = verification.location_id
        and forecasts.valid_at = verification.valid_at
)

select
    *,
    forecast_temperature_2m - reference_temperature_2m as temperature_error,
    abs(forecast_temperature_2m - reference_temperature_2m) as temperature_absolute_error,
    pow(forecast_temperature_2m - reference_temperature_2m, 2) as temperature_squared_error,
    forecast_apparent_temperature - reference_apparent_temperature as apparent_temperature_error,
    abs(
        forecast_apparent_temperature - reference_apparent_temperature
    ) as apparent_temperature_absolute_error,
    forecast_relative_humidity_2m - reference_relative_humidity_2m as relative_humidity_error,
    abs(
        forecast_relative_humidity_2m - reference_relative_humidity_2m
    ) as relative_humidity_absolute_error,
    forecast_precipitation - reference_precipitation as precipitation_error,
    abs(forecast_precipitation - reference_precipitation) as precipitation_absolute_error,
    pow(forecast_precipitation - reference_precipitation, 2) as precipitation_squared_error,
    forecast_rain - reference_rain as rain_error,
    abs(forecast_rain - reference_rain) as rain_absolute_error,
    forecast_snowfall - reference_snowfall as snowfall_error,
    abs(forecast_snowfall - reference_snowfall) as snowfall_absolute_error,
    forecast_cloud_cover - reference_cloud_cover as cloud_cover_error,
    abs(forecast_cloud_cover - reference_cloud_cover) as cloud_cover_absolute_error,
    forecast_pressure_msl - reference_pressure_msl as pressure_msl_error,
    abs(forecast_pressure_msl - reference_pressure_msl) as pressure_msl_absolute_error,
    forecast_visibility - reference_visibility as visibility_error,
    abs(forecast_visibility - reference_visibility) as visibility_absolute_error,
    forecast_wind_speed_10m - reference_wind_speed_10m as wind_speed_10m_error,
    abs(forecast_wind_speed_10m - reference_wind_speed_10m) as wind_speed_10m_absolute_error,
    least(
        abs(forecast_wind_direction_10m - reference_wind_direction_10m),
        360 - abs(forecast_wind_direction_10m - reference_wind_direction_10m)
    ) as wind_direction_10m_absolute_error,
    forecast_wind_gusts_10m - reference_wind_gusts_10m as wind_gusts_10m_error,
    abs(forecast_wind_gusts_10m - reference_wind_gusts_10m) as wind_gusts_10m_absolute_error,
    forecast_uv_index - reference_uv_index as uv_index_error,
    abs(forecast_uv_index - reference_uv_index) as uv_index_absolute_error,
    forecast_weather_code = reference_weather_code as weather_code_exact_match,
    forecast_is_day = reference_is_day as day_night_match,
    case
        when reference_precipitation is null then null
        else reference_precipitation >= 0.1
    end as precipitation_occurred,
    case
        when forecast_precipitation_probability is null
            or reference_precipitation is null
            then null
        when reference_precipitation >= 0.1
            then pow(forecast_precipitation_probability / 100.0 - 1.0, 2)
        else pow(forecast_precipitation_probability / 100.0, 2)
    end as precipitation_probability_brier_score
from matched
