# Grafana Loki Load Testing

```shell
LOKI_URL=http://local-yusuf@localhost:3100 ./k6 run -o json="writer.json" writer.js
LOKI_URL=http://local-yusuf@localhost:3100 ./k6 run -o json="reader.log" reader.js
```


```shell
docker build --platform linux/amd64 -t yusufs/k6 .
```

## Kubernetes

### Installing K6 Operator

```shell
curl https://raw.githubusercontent.com/grafana/k6-operator/4a09026463a5f02f0c1c54cb873e321fd0b3f14a/bundle.yaml | kubectl apply -f -
```

> To delete:
> 
> curl https://raw.githubusercontent.com/grafana/k6-operator/4a09026463a5f02f0c1c54cb873e321fd0b3f14a/bundle.yaml | kubectl delete -f -

### Manual Apply

**Create Namespace**
```shell
kubectl create namespace k6
```

**Create PVC**

This is to save the result of K6 test in file then generate report from that result.

```shell
kubectl -n k6 apply -f kube-pvc.yaml
```

**Create ConfigMap**

```shell
kubectl -n k6 create configmap loki-load-test-file --from-file writer.js --from-file reader.js
```

To delete previous ConfigMap:

```shell
kubectl -n k6 delete configmap loki-load-test-file
```

**Run Test**

```shell
kubectl -n k6 apply -f kube-k6-loki-reader.yaml
kubectl -n k6 apply -f kube-k6-loki-writer.yaml
```


When finished, delete:

```shell
kubectl -n k6 delete -f kube-k6-loki-reader.yaml
kubectl -n k6 delete -f kube-k6-loki-writer.yaml
```

