# Common Dev Config

This repository contains bunch of YAML or any configuration that frequently needed by developer.

It's contains of Docker Compose YAML so we can easily run it in our local machine.


## Rationale

Most of the time, as developer we want to run many of service dependencies: Postgres, Elasticsearch, Kafka, you-name-it.
Installing as real program in your host machine will take a lot of time and configuration.

Docker make this process **a bit** easier but, how if we move from one machine to another? We need a collection of docker-compose.yaml file.

This repository will try to put many Docker Compose with just one common command:

```
docker compose -f service-name.yaml --env-file service-name.env up -d
```

and it should be run with minimal effort of configuration.

But, it needs to be keep data in host volume in case we `down` and `up` the service again.

By default, Docker Compose use `.env` file as stated here https://docs.docker.com/compose/env-file/


## Known Constraint

When we copy the folder, for example `docker-compose/kafka` folder, and run it with the same command `docker compose -f kafka.yaml`
it may return failed because we hard-coded the `container_name` on each YAML file. 
This has the advantages that we can refer to the container by simple name.

If you copy the folder (or cherry-picking files), ensure that you:

* also copy the `.env` file, for example `jaeger.yaml` has `jaeger.env` to set the env variable
* change the `container_name` in the docker compose YAML file to ensure that it won't collides with the default name.
  For example, `jaeger-all-in-one.yaml` has `container_name: jaeger`, change it by adding your service name as prefix or suffix, 
  i.e: your service name `foo` then change it to `foo_jaeger` or `jaeger_foo` as container name.




This repo is for development only