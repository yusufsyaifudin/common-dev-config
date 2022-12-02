# kafka to kafka

Prove that we can send additional Kafka header based on message payload using `meta` in Bloblang.

To try, publish message to topic `in_kafka`:

```shell
echo '{"retry": 50}' | kcat -b localhost:9092 -P -t in_kafka -H 'retry=60'
```

Then in `out_kafka` you will get the payload:

```json
{
	"retry": 50
}
```

and the header something like this:

```json
{
	"kafka_offset": "2",
	"kafka_timestamp_unix": "1669976622",
	"kafka_partition": "1",
	"kafka_key": "",
	"kafka_topic": "in_kafka",
	"kafka_lag": "0",
	"retry_header": "60",
	"retry": "50"
}
```

You will get `retry_header` in Kafka header topic `out_kafka` from header we sent in `in_kafka` topic.
And header `retry` from json payload value.
