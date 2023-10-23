# Common Dev Config

Welcome to the Common Dev Config repository! This collection of YAML and configuration files is designed to simplify the development process for developers.


## Rationale

As developers, we often need to run various service dependencies like Postgres, Elasticsearch, Kafka, and many others. Installing these services directly on your local machine can be time-consuming and requires a lot of configuration.

Docker make this process **a bit** easier but, how if we move from one machine to another? We need a collection of docker-compose.yaml file.

This repository aims to provide a solution for this issue by offering multiple Docker Compose configurations with just one simple command:

```
docker compose -f service-name.yaml --env-file service-name.env up -d
```

With minimal configuration effort, you can get your services up and running. Additionally, it ensures that data is stored in host volumes to prevent data loss when stopping and starting services.

But, it needs to be keep data in host volume in case we `down` and `up` the service again.

By default, Docker Compose use `.env` file as stated here https://docs.docker.com/compose/env-file/


## Known Constraint

Please be aware that this repository is intended for development purposes only. When copying a folder, such as `docker-compose/kafka`, and running it with the same command `docker compose -f kafka.yaml`, it may fail. This is because the container_name is hard-coded in each YAML file. This hard-coding has its advantages, as it allows you to refer to the container by a simple name.

If you decide to copy a folder (or cherry-pick files), make sure you do the following:

* also copy the `.env` file, for example `jaeger.yaml` has `jaeger.env` to set the env variable
* change the `container_name` in the docker compose YAML file to ensure that it won't collides with the default name.
  For example, `jaeger-all-in-one.yaml` has `container_name: jaeger`, change it by adding your service name as prefix or suffix, 

  i.e: your service name `foo` then change it to `foo_jaeger` or `jaeger_foo` as container name.


This repository is a valuable tool for streamlining the development process, making your life as a developer easier.