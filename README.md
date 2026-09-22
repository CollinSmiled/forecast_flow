# Forecast Flow

A weather app for Asian cities backed by an event-driven data platform.

The app serves the latest forecast. The data platform keeps every forecast run
so we can study revisions, lead time, model performance, and accuracy.

```text
Open-Meteo -> Go ingestion -> Kafka -> hot path -> PostgreSQL -> Go API -> Flutter
                                  `-> cold path -> BigQuery -> dbt -> planned Power BI
```

The key rule is simple: a forecast has both a run time (`forecast_run_at`) and
the time it predicts (`valid_at`). New runs do not overwrite old ones.

## Implemented stack

Flutter, Dart, Go, Kafka, PostgreSQL, Open-Meteo, BigQuery batch loading, dbt
warehouse models, and Docker Compose.

Automated dbt deployment and Power BI remain unfinished analytics work.

## Status

The working hot path includes scheduled ingestion, Kafka delivery, PostgreSQL
materialization, a Go API, and a Flutter Android/iOS client with city search,
recent cities, dynamic scenes, and current, hourly, and daily forecasts.

The runnable cold-path service consumes both forecast topics in configurable
micro-batches and uses atomic BigQuery load jobs. The initial dbt project
deduplicates raw events, expands hourly forecasts, and exposes operational,
model-run, and forecast-revision marts. Automated deployment and reporting are
not finished yet.

## Cold-path configuration

The cold-path service defaults to one-hour batches with at most 500 Kafka
records. It uses Google Application Default Credentials and requires
`GOOGLE_CLOUD_PROJECT`. Dataset and table DDL lives under
`warehouse/bigquery/ddl`.

The downstream dbt project lives under `warehouse/dbt`. It initially uses
views so it is compatible with BigQuery Sandbox.

After authenticating locally and creating the tables, run:

```text
go run ./apps/coldpath
```
