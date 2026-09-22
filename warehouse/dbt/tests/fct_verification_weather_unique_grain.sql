select
    location_id,
    valid_at,
    count(*) as row_count
from {{ ref('fct_verification_weather') }}
group by location_id, valid_at
having count(*) > 1
