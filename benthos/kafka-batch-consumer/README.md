# Kafka Batch Consumer

## Features

* [x] Benthos will Batch all consumed message each 5seconds, then send to HTTP Client output.
* [x] All consumed output will still be printed as one message.

## Required Services

* [x] Benthos: echo-server -> This to log our batch and group message in separate service.
* [x] Kafka Broker
* [x] Jaeger

## Required Environment Variables

```shell
HTTP_SERVICE_URL=http://HostIP:4196/post
KAFKA_BROKER_1=kafka:9092
KAFKA_BROKER_2=kafka:9092
KAFKA_BROKER_3=kafka:9092
KAFKA_TOPIC=benthos_batching
KAFKA_CONSUMER_GROUP=my-consumer-group
JAEGER_AGENT_URL=HostIP:6831
JAEGER_COLLECTOR_URL=http://HostIP:14268/api/traces
```

## How to run

```shell
benthos -c config.yaml -r "resources/*.yaml" streams streams/*.yaml
```

Example message payload:

Use `publish_message.sh` script to produce this kind of messages.

```json
{"id":"1","p":1,"status":"sent","msg":"hi-1"}
{"id":"1","p":1,"status":"delivered","msg":"hi-1"}
{"id":"1","p":1,"status":"read","msg":"hi-1"}
{"id":"2","p":2,"status":"sent","msg":"hi-2"}
{"id":"2","p":2,"status":"delivered","msg":"hi-2"}
{"id":"2","p":2,"status":"read","msg":"hi-2"}
```