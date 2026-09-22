with enriched as (
    select
        accuracy.*,
        locations.city,
        locations.administrative_area,
        locations.country,
        locations.country_code,
        locations.timezone
    from {{ ref('fct_model_forecast_accuracy') }} as accuracy
    inner join {{ ref('dim_locations') }} as locations
        on accuracy.location_id = locations.location_id
)

select
    location_id,
    city,
    administrative_area,
    country,
    country_code,
    timezone,
    model_id,
    model_name,
    provider,
    resolution_km,
    lead_time_bucket,
    count(*) as matched_hour_count,
    count(distinct forecast_event_id) as forecast_run_count,
    count(distinct verification_event_id) as verification_event_count,
    min(forecast_run_at) as first_forecast_run_at,
    max(forecast_run_at) as latest_forecast_run_at,
    min(valid_at) as first_valid_at,
    max(valid_at) as latest_valid_at,
    max(verification_retrieved_at) as latest_verification_retrieved_at,
    count(temperature_absolute_error) as temperature_sample_count,
    avg(temperature_error) as temperature_mean_error,
    avg(temperature_absolute_error) as temperature_mae,
    sqrt(avg(temperature_squared_error)) as temperature_rmse,
    avg(apparent_temperature_absolute_error) as apparent_temperature_mae,
    avg(relative_humidity_error) as relative_humidity_mean_error,
    avg(relative_humidity_absolute_error) as relative_humidity_mae,
    count(precipitation_absolute_error) as precipitation_sample_count,
    avg(precipitation_error) as precipitation_mean_error,
    avg(precipitation_absolute_error) as precipitation_mae,
    sqrt(avg(precipitation_squared_error)) as precipitation_rmse,
    avg(rain_absolute_error) as rain_mae,
    avg(snowfall_absolute_error) as snowfall_mae,
    avg(cloud_cover_absolute_error) as cloud_cover_mae,
    avg(pressure_msl_absolute_error) as pressure_msl_mae,
    avg(visibility_absolute_error) as visibility_mae,
    avg(wind_speed_10m_absolute_error) as wind_speed_10m_mae,
    avg(wind_direction_10m_absolute_error) as wind_direction_10m_mae,
    avg(wind_gusts_10m_absolute_error) as wind_gusts_10m_mae,
    avg(uv_index_absolute_error) as uv_index_mae,
    safe_divide(
        countif(weather_code_exact_match),
        count(weather_code_exact_match)
    ) as weather_code_exact_match_rate,
    safe_divide(
        countif(day_night_match),
        count(day_night_match)
    ) as day_night_match_rate,
    avg(precipitation_probability_brier_score) as precipitation_probability_brier_score
from enriched
group by
    location_id,
    city,
    administrative_area,
    country,
    country_code,
    timezone,
    model_id,
    model_name,
    provider,
    resolution_km,
    lead_time_bucket
