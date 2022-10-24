# Docker Compose

## PORT LIST

| Service Name         | Container name | FILE                             | HOST PORT:DOCKER PORT |
|----------------------|----------------|----------------------------------|-----------------------|
| Elasticsearch        | es             | elasticsearch/elasticsearch.yaml | 9200:9200, 9300:9300  |
| Kafka Broker         | kafka          | kafka/kafka.yaml                 | 9092:9092             |
| Kafka UI             | kafkaui        | kafka/kafka.yaml                 | 8080:8080             |
| MongoDB              | mongodb        | mongodb/mongodb.yaml             | 27017:27017           |
| MongoExpress (UI)    | mongoexpress   | mongodb/mongodb.yaml             | 8081:8081             |
| Postgres             | postgres       | postgres/postgres.yaml           | 5433:5432             |
| PgAdmin 4            | pgadmin        | postgres/postgres.yaml           | 5050:80               |


