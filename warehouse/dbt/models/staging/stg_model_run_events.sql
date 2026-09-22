with source_events as (
    select *
    from {{ source('forecast_raw', 'model_run_events') }}
)

select
    event_id,
    event_type,
    schema_version,
    location_id,
    model_id,
    json_value(payload, '$.data.model.model_name') as model_name,
    provider,
    safe_cast(
        json_value(payload, '$.data.model.resolution_km') as float64
    ) as resolution_km,
    forecast_run_at,
    retrieved_at,
    safe_cast(json_value(payload, '$.occurred_at') as timestamp) as event_occurred_at,
    json_query(payload, '$.data.hourly') as hourly_forecasts_json,
    kafka_topic,
    kafka_partition,
    kafka_offset,
    kafka_key,
    kafka_timestamp,
    ingested_at,
    payload
from source_events
qualify row_number() over (
    partition by event_id
    order by ingested_at desc, kafka_partition desc, kafka_offset desc
) = 1
