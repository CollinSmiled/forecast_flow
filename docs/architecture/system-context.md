# System context

Forecast Flow has two views of the same forecast data:

- The weather app needs the latest forecast quickly.
- Analytics needs every version of every forecast.

```text
Open-Meteo -> ingestion -> Kafka -> hot path -> PostgreSQL -> API -> Flutter
                              `-> cold path -> BigQuery -> dbt -> planned Power BI
```

The hot path is implemented. The runnable cold-path service includes row
mapping, three-topic Kafka micro-batching, and atomic BigQuery load jobs. The
initial dbt layers and marts are implemented as BigQuery views. Automated dbt
deployment and Power BI reporting remain planned work.

Forecast verification uses a separate `weather.verification` event stream and
`forecast_raw.verification_weather_events` table. Its reference values are
historical reanalysis, which combines measurements and model estimates; they
must not be presented as direct weather-station observations. The event
contract, Open-Meteo retrieval, daily timezone-aware scheduler, Kafka
publishing, and raw infrastructure are implemented. The scheduler uses a
configurable safety lag (seven days by default) so reanalysis data has time to
become available. The cold path loads these events into BigQuery and dbt
exposes a deduplicated staging view plus typed hourly reference rows. Accuracy
marts remain planned work.

The hourly forecast grain is:

```text
one location x one model x one forecast run x one valid hour
```

Its identity includes `location_id`, `model_id`, `forecast_run_at`, and
`valid_at`. Lead time is the difference between the last two timestamps.

Verification weather has one row per location and valid hour. Accuracy models
will join this grain to forecast facts using `location_id` and `valid_at` while
retaining the verification source and retrieval time.
