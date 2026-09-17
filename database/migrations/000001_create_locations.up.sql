CREATE TABLE public.locations (
    location_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    open_meteo_location_id BIGINT NOT NULL,

    city TEXT NOT NULL,
    country TEXT NOT NULL,
    country_code TEXT NOT NULL,

    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    timezone TEXT NOT NULL,

    elevation DOUBLE PRECISION,
    population BIGINT,
    administrative_area TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_locations_open_meteo_location_id
        UNIQUE (open_meteo_location_id),

    CONSTRAINT chk_locations_country_code
        CHECK (
            char_length(country_code) = 2
            AND country_code = upper(country_code)
        ),

    CONSTRAINT chk_locations_latitude
        CHECK (latitude BETWEEN -90 AND 90),

    CONSTRAINT chk_locations_longitude
        CHECK (longitude BETWEEN -180 AND 180),

    CONSTRAINT chk_locations_population
        CHECK (population IS NULL OR population >= 0)
);

CREATE INDEX idx_locations_city_search
    ON public.locations (lower(city) text_pattern_ops);

CREATE INDEX idx_locations_country_code
    ON public.locations (country_code);