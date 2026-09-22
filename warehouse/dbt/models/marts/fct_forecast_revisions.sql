select
    event_id,
    location_id,
    model_id,
    model_name,
    provider,
    forecast_run_at,
    valid_at,
    lead_time_hours,
    temperature_2m,
    precipitation_probability,
    lag(forecast_run_at) over forecast_sequence as previous_forecast_run_at,
    lag(temperature_2m) over forecast_sequence as previous_temperature_2m,
    temperature_2m
        - lag(temperature_2m) over forecast_sequence
        as temperature_revision,
    lag(precipitation_probability) over forecast_sequence
        as previous_precipitation_probability,
    precipitation_probability
        - lag(precipitation_probability) over forecast_sequence
        as precipitation_probability_revision
from {{ ref('int_model_run_hourly_forecasts') }}
window forecast_sequence as (
    partition by location_id, model_id, valid_at
    order by forecast_run_at, retrieved_at, event_id
)
