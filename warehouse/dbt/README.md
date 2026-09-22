# Forecast Flow dbt warehouse

This project transforms immutable Kafka event payloads from `forecast_raw`
and operational reference snapshots from `forecast_reference` into documented,
query-friendly BigQuery views.

## Data flow

```text
forecast_raw
  -> forecast_staging
  -> forecast_intermediate
  -> forecast_marts
  -> BI and analysis

forecast_reference
  -> forecast_staging
  -> forecast_marts
```

The `forecast_raw` dataset is owned by the Go cold-path service, while the
`forecast_reference` dataset is refreshed by the location-sync worker. dbt
owns all three downstream datasets. Models are views so the project works in
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

Pull requests and pushes to `main` run a credential-free `dbt parse` check in
GitHub Actions. This validates project configuration, model Jinja and YAML
before any live BigQuery deployment is attempted. A live `dbt build` remains
the check for BigQuery SQL execution and data tests.

The star-schema dimensions provide one descriptive row per weather model in
`forecast_marts.dim_forecast_models` and one row per supported city in
`forecast_marts.dim_locations`. Relationship tests verify that forecast facts
reference valid dimension records.

## Verification and accuracy

The verification pipeline treats Open-Meteo historical reanalysis as a
reference, not as direct weather-station observations. Its main models are:

- `stg_verification_weather_events`: deduplicated event envelopes.
- `int_verification_hourly_weather`: typed hourly rows retaining every event
  version.
- `fct_verification_weather`: latest reference version per location and valid
  hour.
- `fct_model_forecast_accuracy`: model forecasts matched to reference weather
  by `location_id` and `valid_at`, with per-hour errors.
- `agg_model_accuracy`: dashboard-ready metrics by model, city, and lead-time
  bucket.

Temperature and precipitation include mean error, MAE, and RMSE inputs. The
hourly accuracy fact also exposes errors for humidity, pressure, visibility,
wind, cloud cover, UV, and weather code. Precipitation probability is evaluated
with a Brier score using 0.1 mm as the event threshold.

Accuracy models can validly contain zero rows while the newest forecasts wait
for delayed reanalysis. With the default seven-day verification lag, newly
collected forecasts begin matching approximately one week later. Because the
models are views, no dbt rerun is required when those source rows arrive.
