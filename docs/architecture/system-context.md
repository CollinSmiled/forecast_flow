# System context

Forecast Flow has two views of the same forecast data:

- The weather app needs the latest forecast quickly.
- Analytics needs every version of every forecast.

```text
Open-Meteo -> ingestion -> Kafka -> hot path  -> PostgreSQL -> API -> web
                              `-> cold path -> BigQuery -> dbt -> Power BI
```

The hourly forecast grain is:

```text
one location x one model x one forecast run x one valid hour
```

Its identity includes `location_id`, `model_id`, `forecast_run_at`, and
`valid_at`. Lead time is the difference between the last two timestamps.
