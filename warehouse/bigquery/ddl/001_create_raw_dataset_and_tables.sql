CREATE SCHEMA IF NOT EXISTS forecast_raw
OPTIONS (
    location = 'asia-southeast2',
    description = 'Immutable Kafka events received from the forecast platform'
);

CREATE TABLE IF NOT EXISTS forecast_raw.operational_forecast_events (
    event_id STRING NOT NULL,
    event_type STRING NOT NULL,
    schema_version INT64 NOT NULL,

    location_id INT64 NOT NULL,
    source STRING NOT NULL,
    retrieved_at TIMESTAMP NOT NULL,

    kafka_topic STRING NOT NULL,
    kafka_partition INT64 NOT NULL,
    kafka_offset INT64 NOT NULL,
    kafka_key STRING NOT NULL,
    kafka_timestamp TIMESTAMP NOT NULL,

    payload STRING NOT NULL,
    ingested_at TIMESTAMP NOT NULL
)
PARTITION BY DATE(retrieved_at)
CLUSTER BY location_id, event_id;

CREATE TABLE IF NOT EXISTS forecast_raw.model_run_events (
    event_id STRING NOT NULL,
    event_type STRING NOT NULL,
    schema_version INT64 NOT NULL,

    location_id INT64 NOT NULL,
    model_id STRING NOT NULL,
    provider STRING NOT NULL,
    forecast_run_at TIMESTAMP NOT NULL,
    retrieved_at TIMESTAMP NOT NULL,

    kafka_topic STRING NOT NULL,
    kafka_partition INT64 NOT NULL,
    kafka_offset INT64 NOT NULL,
    kafka_key STRING NOT NULL,
    kafka_timestamp TIMESTAMP NOT NULL,

    payload STRING NOT NULL,
    ingested_at TIMESTAMP NOT NULL
)
PARTITION BY DATE(forecast_run_at)
CLUSTER BY location_id, model_id, event_id;
