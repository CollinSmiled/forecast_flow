select *
from {{ ref('fct_model_forecast_accuracy') }}
where temperature_absolute_error < 0
    or temperature_squared_error < 0
    or precipitation_absolute_error < 0
    or precipitation_squared_error < 0
    or wind_direction_10m_absolute_error < 0
    or wind_direction_10m_absolute_error > 180
    or precipitation_probability_brier_score < 0
    or precipitation_probability_brier_score > 1
