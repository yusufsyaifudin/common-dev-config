# Grafana Loki Microservice Deployment Setup

Source codes in this directory are influenced from [this repository](https://github.com/saidsef/grafana-loki-on-k8s/tree/cd3cd72e86f72e99986c250114a3324a70b83561),
but since it just deploy using Monolith Deployment, this directory contains separate deployment of each Grafana Loki Components:


| Loki Components | Mode                    | Requests     | Limits     | Replicas | Volume Claim | 
|-----------------|-------------------------|--------------|------------|----------|--------------|
| compactor       | Stateful                | 200m/200Mi   | 1000m/1Gi  | 1        | 1Gi          |
| distributor     | Stateless               | 200m/500Mi   | 1000m/1Gi  | 3        |              |
| index-gateway   | Stateful                | 500m/500Mi   | 1000m/1Gi  | 1        | 1Gi          |
| ingester        | Stateful                | 200m/200Mi   | 1000m/1Gi  | 3        | 2Gi          |
| querier         | Stateless               | 200m/500Mi   | 1000m/1Gi  | 3        |              |
| query-frontend  | Stateless               | 200m/200Mi   | 500m/500Mi | 2        |              |
| query-scheduler | Stateless               | 500m/500Mi   | 2000m/1Gi  | 1        |              |
| ruler           | Stateful                | 100m/200Mi   | 500m/500Mi | 1        | 500Mi        |
|                 |                         |              |            |          |              |
| **Total**       | 4 Stateful, 4 Stateless | 2100m/2800Mi | 8000m/7Gi  | 15 pods  | 4.5 GB       |


* Compactor: [Run the compactor as a singleton (a single instance).](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/operations/storage/retention.md?plain=1#L19)
* Distributor: [The distributor service is responsible for handling incoming streams by clients. It's the first stop in the write path for log data. Once the distributor receives a set of streams, each stream is validated for correctness and to ensure that it is within the configured tenant (or global) limits. Valid chunks are then split into batches and sent to multiple ingesters in parallel.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/fundamentals/architecture/components/_index.md?plain=1#L12-L17)
  Because of this, we need to run minimum 3 replicas so, we will get higher throughput of writer. Since it act as buffer, we need to give a higher RAM to make sure all logs is retrieved.
* Index Gateway: Index gateway makes Querier component can be deployed as Stateless. All Querier will read index from this and by default it has `cache_ttl: 24h`. 
  Because of this, we need to deploy it with Statefulset and set with higher resources. References:
  * https://github.com/grafana/helm-charts/issues/611
  * https://github.com/grafana/loki/issues/6369
* Ingester: [Ingesters temporarily store data in memory. In the event of a crash, there could be data loss.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/operations/storage/wal.md?plain=1#L8)
  We set `replay_memory_ceiling: 750MB` because we set maximum limit 1Gb memory. We set Volume Claim twice as limit memory to ensure all WAL data can persist into persistent storage during shutdown.
* Querier: [See this document about queriers.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/operations/autoscaling_queriers.md?plain=1#L109)
  It is stateless component that run the actual read query from the Grafana. But, since we deploy using Microservice architecture it will read the index from the Index Gateway.
  To accepts higher read throughput, we need 3 replicas.
* Query Frontend: This query frontend is to minimize bottleneck on the read path, [read more here.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/configuration/query-frontend.md)
  Since this is the first point of the read path (when we query from Grafana dashboard), then we need two replicas to ensure that query is run smoothly.
* Query Scheduler: [Read more here.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/operations/scalability.md)   
  As stated in the documentation, that it may run frequent garbage collector, so we start 500 milli-core CPU with 500 Megabytes RAM, with upper limit up-to 2 cores CPU and 1GB RAM.
* Ruler: [The ruler is responsible for continually evaluating a set of configurable queries and performing an action based on the result.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/rules/_index.md?plain=1#L11)
  Because this components is _only_ for sending metrics of Grafana Loki internal process, then we allocate only small resources.
  This also need Persisent Volume as stated in the documentation (link below), 500 MB is enough to save rulers' WAL. [WALs should not grow
  excessively large due to truncation.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/operations/recording-rules.md?plain=1#L47-L48)
  * [For those looking to get started with metrics and alerts based on logs.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/rules/_index.md?plain=1#L134)
  * [Loki's implementation of recording rules largely reuses Prometheus' code.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/operations/recording-rules.md?plain=1#L12)
  * [It is recommended that you run the `rulers` using `StatefulSets`. The `ruler` will write its WAL files to persistent storage,
    so a `Persistent Volume` should be utilised.](https://github.com/grafana/loki/blob/v2.8.3/docs/sources/operations/recording-rules.md?plain=1#L65)

> References:
> * https://community.grafana.com/t/grafana-loki-stateful-vs-stateless-components/100237
> * https://community.grafana.com/t/memberlist-join-members-configuration-for-read-write-deployments/100086


## Prerequisites

I've tested this command using [Minikube v1.31.1](https://github.com/kubernetes/minikube/releases/tag/v1.31.1):

```shell
minikube start --memory 4096 --cpus 2 --ports=30000:30000 --kubernetes-version=v1.27.3
```

> Links:
> * [Minikube start with more memory](https://www.shellhacks.com/minikube-start-with-more-memory-cpus/)


After you have the Kubernetes cluster ready, then follow these step.

## Running Minio

You can skip this if you want to use AWS S3 or any object storage.

