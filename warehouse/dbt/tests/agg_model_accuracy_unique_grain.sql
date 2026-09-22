select
    location_id,
    model_id,
    lead_time_bucket,
    count(*) as row_count
from {{ ref('agg_model_accuracy') }}
group by location_id, model_id, lead_time_bucket
having count(*) > 1
