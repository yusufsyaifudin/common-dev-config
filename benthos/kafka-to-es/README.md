# Kafka to Elasticsearch

## Features

* [x] If persist to ES failed -> go to another Kafka topic (DLQ)

## Required Services

* [x] Elasticsearch
* [x] Kafka Broker
* [x] Jaeger

## Required Environment Variables

```shell
ES_URL="https://postman-echo.com/post"
KAFKA_BROKER_1=kafka:9092
KAFKA_BROKER_2=kafka:9092
KAFKA_BROKER_3=kafka:9092
KAFKA_CONSUMER_GROUP=my-consumer-group-es
KAFKA_TOPIC=my-topic-es
KAFKA_TOPIC_DLQ=my-topic-es-dlq
```

## How to run

```shell
benthos -c config.yaml -r "resources/*.yaml" streams streams/*.yaml
```