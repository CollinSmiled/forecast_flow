select
    forecast_event_id,
    valid_at,
    count(*) as row_count
from {{ ref('fct_model_forecast_accuracy') }}
group by forecast_event_id, valid_at
having count(*) > 1
