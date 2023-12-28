# Log Generator


## Build Docker Image

```shell
docker buildx build --push --platform linux/amd64,linux/arm64 -t yusufs/log-gen:latest .
```

## REST API

### Generate log 1MBps for 1 minute

```shell
curl -X GET "http://localhost:4000/generate-log?bytes_per_second=1mb&duration=1m&start_time=2023-11-06T21:20:00%2b07:00&request_id=uuid"
```

Response example:

```shell
{
    "duration": 60000000000,
    "durationHumanize": "1m0s",
    "maxBytesPerSecond": 1000000,
    "maxBytesPerSecondHumanize": "1.0 MB",
    "requestID": "123",
    "timeEnd": "2023-11-06T21:21:02+07:00",
    "timeStart": "2023-11-06T21:20:01+07:00"
}
```

### Get Statistic of the request ID

```shell
curl -X GET " http://localhost:4000/generate-log-status/uuid"
```

Response example:

```json lines
{
  "timeStarted": "2023-11-06T21:20:01+07:00",
  "timeEnded": "2023-11-06T21:21:01+07:00",
  "totalLines": 54882,
  "totalBytes": 61036472,
  "totalBytesHumanize": "61 MB",
  "bytesPerSecond": [
    {
      "date": "2023-11-06T21:20:01+07:00",
      "totalLines": 900,
      "totalBytes": 1000835,
      "totalBytesHumanize": "1.0 MB"
    },
    // line omitted for easier read
  ]
}
```
