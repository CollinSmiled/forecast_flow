CREATE TABLE IF NOT EXISTS forecast_raw.verification_weather_events (
    event_id STRING NOT NULL,
    event_type STRING NOT NULL,
    schema_version INT64 NOT NULL,

    location_id INT64 NOT NULL,
    source STRING NOT NULL,
    reference_kind STRING NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    retrieved_at TIMESTAMP NOT NULL,

    kafka_topic STRING NOT NULL,
    kafka_partition INT64 NOT NULL,
    kafka_offset INT64 NOT NULL,
    kafka_key STRING NOT NULL,
    kafka_timestamp TIMESTAMP NOT NULL,

    payload STRING NOT NULL,
    ingested_at TIMESTAMP NOT NULL
)
PARTITION BY DATE(period_start)
CLUSTER BY location_id, source, event_id;
