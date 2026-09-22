select
    event_id,
    valid_at,
    count(*) as row_count
from {{ ref('fct_operational_hourly_forecasts') }}
group by event_id, valid_at
having count(*) > 1
