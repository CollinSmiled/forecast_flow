select
    event_id,
    valid_at,
    count(*) as row_count
from {{ ref('int_verification_hourly_weather') }}
group by event_id, valid_at
having count(*) > 1
