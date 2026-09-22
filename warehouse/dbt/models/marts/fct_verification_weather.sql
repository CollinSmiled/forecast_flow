with ranked as (
    select
        event_id,
        location_id,
        valid_at,
        date(valid_at) as verification_date_utc,
        source,
        reference_kind,
        retrieved_at,
        verification_timezone,
        period_start,
        period_end,
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
        ingested_at,
        count(*) over (
            partition by location_id, valid_at
        ) as verification_version_count,
        row_number() over (
            partition by location_id, valid_at
            order by retrieved_at desc, ingested_at desc, event_id desc
        ) as version_rank
    from {{ ref('int_verification_hourly_weather') }}
)

select * except (version_rank)
from ranked
where version_rank = 1
