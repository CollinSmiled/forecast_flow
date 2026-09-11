# Forecast Flow

A weather app for Asian cities backed by an event-driven data platform.

The app serves the latest forecast. The data platform keeps every forecast run
so we can study revisions, lead time, model performance, and accuracy.

```text
Open-Meteo -> Go -> Kafka
                       |-> PostgreSQL -> API -> web app
                       `-> BigQuery -> dbt -> Power BI
```

The key rule is simple: a forecast has both a run time (`forecast_run_at`) and
the time it predicts (`valid_at`). New runs do not overwrite old ones.

## Stack

Go, Kafka, PostgreSQL, BigQuery, dbt, Power BI, React, TypeScript, and Docker.

## Status

Building the project one working increment at a time. Currently setting up the
repository and documenting the architecture.
