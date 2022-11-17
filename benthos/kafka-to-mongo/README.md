# Kafka to MongoDB

## Features

* [x] If persist to MongoDB failed -> go to another Kafka topic (DLQ)
* [x] If DLQ queue still failed persist to MongoDB -> then trying to push in Flock as message (you can send it to another system as long as it has HTTP endpoint)

## Required Services

* [x] MongoDB
* [x] Kafka Broker
* [x] Jaeger

## Required Environment Variables

```shell
MONGO_URL=http://hostIP:9200
MONGO_DB=kafka_messages
MONGO_USERNAME=root
MONGO_PASSWORD=password

KAFKA_BROKER_1=1hostIP:9092
KAFKA_BROKER_2=hostIP:9092
KAFKA_BROKER_3=hostIP:9092
KAFKA_CONSUMER_GROUP=my-consumer-group5
KAFKA_TOPIC=benthos-mongo
KAFKA_TOPIC_DLQ=benthos-mongo-dlq
```

## How to run

```shell
benthos -c config.yaml -r "resources/*.yaml" streams streams/*.yaml
```

Then publish message with this structure:

```json
{
  "header": {
    "target_collection": "service-1"
  },
  "body": {
    "key": "data-",
    "another_key": {
      "foo": "bar"
    }
  }
}
```