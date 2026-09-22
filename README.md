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

The runnable cold-path service consumes operational forecasts, model runs, and
verification weather in configurable micro-batches and uses atomic BigQuery
load jobs. The dbt project deduplicates raw events, expands hourly data, and
exposes operational, model-run, revision, verification, and model-accuracy
marts. A dashboard-ready accuracy summary groups results by model, city, and
forecast lead time. Automated deployment and reporting are not finished yet.

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

The current PostgreSQL location catalog is synchronized separately into the
`forecast_reference.locations` BigQuery table. After applying the reference
DDL, run a one-shot synchronization from the repository root:

```text
go run ./apps/locationsync
```

For continuous local analytics, the optional Compose services run the Kafka
cold-path consumer, synchronize locations, and retrieve delayed verification
weather:

```text
docker compose --profile analytics up -d coldpath location-sync verification-scheduler
```

Location synchronization runs immediately and then hourly by default.
Verification retrieval runs immediately and then daily, selecting seven local
calendar days ago for each city so Open-Meteo reanalysis has time to become
available. Configure these schedules with `LOCATION_SYNC_INTERVAL`,
`VERIFICATION_POLL_INTERVAL`, and `VERIFICATION_LAG_DAYS`.

dbt models are BigQuery views, so new matching source rows appear automatically;
`dbt build` is only required when deploying or testing model changes. Accuracy
views remain empty until stored model forecasts and delayed verification data
share the same `location_id` and `valid_at`.
