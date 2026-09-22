with model_runs as (
    select distinct
        event_id,
        model_id,
        model_name,
        provider,
        resolution_km,
        forecast_run_at,
        retrieved_at
    from {{ ref('int_model_run_hourly_forecasts') }}
),

latest_attributes as (
    select
        model_id,
        model_name,
        provider,
        resolution_km
    from model_runs
    qualify row_number() over (
        partition by model_id
        order by forecast_run_at desc, retrieved_at desc, event_id desc
    ) = 1
),

run_summary as (
    select
        model_id,
        min(forecast_run_at) as first_forecast_run_at,
        max(forecast_run_at) as latest_forecast_run_at,
        count(*) as forecast_run_count
    from model_runs
    group by model_id
)

select
    latest_attributes.model_id,
    latest_attributes.model_name,
    latest_attributes.provider,
    latest_attributes.resolution_km,
    run_summary.first_forecast_run_at,
    run_summary.latest_forecast_run_at,
    run_summary.forecast_run_count
from latest_attributes
inner join run_summary using (model_id)
