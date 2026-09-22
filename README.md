# Forecast Flow

A weather app for Asian cities backed by an event-driven data platform.

The app serves the latest forecast. The data platform keeps every forecast run
so we can study revisions, lead time, model performance, and accuracy.

```text
Open-Meteo -> Go ingestion -> Kafka -> hot path -> PostgreSQL -> Go API -> Flutter
                                  `-> cold path -> BigQuery -> planned dbt/Power BI
```

The key rule is simple: a forecast has both a run time (`forecast_run_at`) and
the time it predicts (`valid_at`). New runs do not overwrite old ones.

## Implemented stack

Flutter, Dart, Go, Kafka, PostgreSQL, Open-Meteo, BigQuery batch loading, and
Docker Compose.

BigQuery cloud provisioning, dbt, and Power BI remain unfinished analytics
work.

## Status

The working hot path includes scheduled ingestion, Kafka delivery, PostgreSQL
materialization, a Go API, and a Flutter Android/iOS client with city search,
recent cities, dynamic scenes, and current, hourly, and daily forecasts.

The runnable cold-path service consumes both forecast topics in configurable
micro-batches and uses atomic BigQuery load jobs. Cloud provisioning,
deployment, and analytics models are not finished yet.

## Cold-path configuration

The cold-path service defaults to one-hour batches with at most 500 Kafka
records. It uses Google Application Default Credentials and requires
`GOOGLE_CLOUD_PROJECT`. Dataset and table DDL lives under
`warehouse/bigquery/ddl`.

After authenticating locally and creating the tables, run:

```text
go run ./apps/coldpath
```
