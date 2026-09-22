# Forecast Flow

A weather app for Asian cities backed by an event-driven data platform.

The app serves the latest forecast. The data platform keeps every forecast run
so we can study revisions, lead time, model performance, and accuracy.

```text
Open-Meteo -> Go ingestion -> Kafka -> hot path -> PostgreSQL -> Go API -> Flutter
                                  `-> planned cold path -> BigQuery -> dbt -> Power BI
```

The key rule is simple: a forecast has both a run time (`forecast_run_at`) and
the time it predicts (`valid_at`). New runs do not overwrite old ones.

## Implemented stack

Flutter, Dart, Go, Kafka, PostgreSQL, Open-Meteo, and Docker Compose.

BigQuery, dbt, and Power BI remain part of the planned analytics cold path.

## Status

The working hot path includes scheduled ingestion, Kafka delivery, PostgreSQL
materialization, a Go API, and a Flutter Android/iOS client with city search,
recent cities, dynamic scenes, and current, hourly, and daily forecasts.

The versioned forecast-run event, BigQuery row mapping and writer, cold-path
event processor, and dual-topic Kafka consumer exist. The runnable cold-path
service, deployment, and analytics models are not finished yet.
