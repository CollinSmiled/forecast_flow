select
    location_id,
    open_meteo_location_id,
    city,
    administrative_area,
    country,
    country_code,
    latitude,
    longitude,
    timezone,
    elevation,
    population,
    source_created_at,
    source_updated_at,
    synced_at
from {{ ref('stg_locations') }}
