select *
from {{ ref('agg_model_accuracy') }}
where matched_hour_count < 1
    or temperature_mae < 0
    or temperature_rmse < 0
    or precipitation_mae < 0
    or precipitation_rmse < 0
    or weather_code_exact_match_rate < 0
    or weather_code_exact_match_rate > 1
    or day_night_match_rate < 0
    or day_night_match_rate > 1
    or precipitation_probability_brier_score < 0
    or precipitation_probability_brier_score > 1
