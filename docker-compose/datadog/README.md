# Datadog Agent


```shell
docker run -d --cgroupns host --pid host --name dd-agent -v /var/run/docker.sock:/var/run/docker.sock:ro -v /proc/:/host/proc/:ro -v /sys/fs/cgroup/:/host/sys/fs/cgroup:ro -e DD_SITE=<DATADOG_SITE> -e DD_API_KEY=<DATADOG_API_KEY> gcr.io/datadoghq/agent:7

```


```shell
docker run -d --cgroupns host --pid host --name dd-agent -e DD_HOSTNAME=ddagent -e DD_SITE=datadoghq.com -e DD_API_KEY=xxx -p 8125:8125/udp -p 8126:8126 gcr.io/datadoghq/agent:7

```