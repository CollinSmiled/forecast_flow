#!/bin/sh

set -eu

create_topic() {
    /opt/kafka/bin/kafka-topics.sh \
        --bootstrap-server kafka:19092 \
        --create \
        --if-not-exists \
        --topic "$1" \
        --partitions 3 \
        --replication-factor 1
}

create_topic forecast.latest
create_topic forecast.runs
