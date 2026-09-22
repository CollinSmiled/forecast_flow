CREATE SCHEMA IF NOT EXISTS forecast_reference
OPTIONS (
    location = 'asia-southeast2',
    description = 'Current reference data synchronized from operational systems'
);

CREATE TABLE IF NOT EXISTS forecast_reference.locations (
    location_id INT64 NOT NULL,
    open_meteo_location_id INT64 NOT NULL,

    city STRING NOT NULL,
    country STRING NOT NULL,
    country_code STRING NOT NULL,

    latitude FLOAT64 NOT NULL,
    longitude FLOAT64 NOT NULL,
    timezone STRING NOT NULL,

    elevation FLOAT64,
    population INT64,
    administrative_area STRING,

    source_created_at TIMESTAMP NOT NULL,
    source_updated_at TIMESTAMP NOT NULL,
    synced_at TIMESTAMP NOT NULL
)
CLUSTER BY country_code, location_id;
