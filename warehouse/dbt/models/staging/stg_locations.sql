with source_locations as (
    select *
    from {{ source('forecast_reference', 'locations') }}
)

select
    location_id,
    open_meteo_location_id,
    trim(city) as city,
    trim(country) as country,
    upper(trim(country_code)) as country_code,
    latitude,
    longitude,
    trim(timezone) as timezone,
    elevation,
    population,
    nullif(trim(administrative_area), '') as administrative_area,
    source_created_at,
    source_updated_at,
    synced_at
from source_locations
qualify row_number() over (
    partition by location_id
    order by source_updated_at desc, synced_at desc, open_meteo_location_id desc
) = 1
