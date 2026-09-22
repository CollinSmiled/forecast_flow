# Forecast Flow dbt warehouse

This project transforms immutable Kafka event payloads from `forecast_raw`
into documented, query-friendly BigQuery views.

## Data flow

```text
forecast_raw
  -> forecast_staging
  -> forecast_intermediate
  -> forecast_marts
  -> BI and analysis
```

The `forecast_raw` dataset is owned by the Go cold-path service. dbt owns all
three downstream datasets. Initial models are views so the project works in
BigQuery Sandbox without DML and does not duplicate stored data.

## Local profile

Copy `profiles.example.yml` to `profiles.yml`. The local profile is ignored by
Git. It uses Application Default Credentials and defaults to project
`forecast-flow-509406` in region `asia-southeast2`.

To target another project for a session:

```powershell
$env:DBT_BIGQUERY_PROJECT = "your-project-id"
```

Like the Olist analytics project, Forecast Flow runs dbt Core through the
Docker Compose tool service. From the repository root, run:

```powershell
docker compose --profile tools build dbt
docker compose --profile tools run --rm dbt debug
docker compose --profile tools run --rm dbt build
```

`dbt build` creates the downstream datasets/views, runs generic tests, and
runs the custom grain tests in dependency order.

The first star-schema dimension, `forecast_marts.dim_forecast_models`, provides
one descriptive row per weather model and is tested against the model-based
fact views.
