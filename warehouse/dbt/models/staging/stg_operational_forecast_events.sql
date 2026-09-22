with source_events as (
    select *
    from {{ source('forecast_raw', 'operational_forecast_events') }}
)

select
    event_id,
    event_type,
    schema_version,
    location_id,
    source,
    retrieved_at,
    safe_cast(json_value(payload, '$.occurred_at') as timestamp) as event_occurred_at,
    json_value(payload, '$.data.timezone') as forecast_timezone,
    json_query(payload, '$.data.current') as current_conditions_json,
    json_query(payload, '$.data.hourly') as hourly_forecasts_json,
    json_query(payload, '$.data.daily') as daily_forecasts_json,
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
