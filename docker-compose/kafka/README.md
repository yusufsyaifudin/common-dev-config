# Kafka

## kafka.yaml

### Includes

* [x] Zookeeper v3.8.0
* [x] Kafka v3.3.1
* [x] Kafka-UI v0.4.0

### How to run
Run with:

```shell
docker compose -f kafka.yaml up
```

### Accessing the Kafka Broker From Host Machine

Add this in your `/etc/hosts` (`sudo nano /etc/hosts`):

```shell
127.0.0.1 kafka
```

Now, using [`kcat`](https://formulae.brew.sh/formula/kcat) you can do this:

```shell
kcat -b kafka:9092 -L
```

And if you want to access the Kafka cluster from your host machine, then use `kafka` as the domain.


### Accessing Kafka Broker from Another Docker Container

Get your current ip:

```shell
ipconfig getifaddr en0
```

> Why using `MY_IP=$(ipconfig getifaddr en0)` is because from [Bitnami Readme](https://github.com/bitnami/containers/blob/75805a6610e49214591a254bd3a6a808faf99c19/bitnami/kafka/README.md?plain=1#L326):
> To connect from an external machine, change localhost above to your host's external IP/hostname and include
> EXTERNAL://0.0.0.0:9093 in KAFKA_CFG_LISTENERS to allow for remote connections.


Then in your another Docker container (compose), you can use this IP to connect.

For example, above command return IP `192.168.0.102`, then in your `KAFKA_ADDRESS` environment in another Docker Container
you can use `192.168.0.102:9092` to connect to this Kafka Broker. `KAFKA_ADDRESS` here is depending your application,
for example in our `jaeger/jaeger.yaml`, it use `KAFKA_BROKER` as environment variable key.
