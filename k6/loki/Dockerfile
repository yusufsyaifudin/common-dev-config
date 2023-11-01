# Build the k6 binary with the extension
FROM golang:1.21-alpine3.18 as builder

RUN apk update && apk add bash make git

RUN go install go.k6.io/xk6/cmd/xk6@v0.9.2

WORKDIR /home/k6

RUN git clone https://github.com/grafana/xk6-loki

# using specific commit because we want reproducible action with exact version!
RUN cd /home/k6/xk6-loki && git checkout 12ba135193ecb17f37d043262f2f145d5b9cf641

RUN rm -rf /home/k6/xk6-loki/Makefile

# Feel free to add other extensions using the '--with ...'.
RUN <<EOF cat >> /home/k6/xk6-loki/Makefile
PWD := $(shell pwd)
GO_FILES := $(shell find . -type f -name "*.go" -print)

.PHONY: run

k6: $(GO_FILES)
	xk6 build \
		--replace "github.com/hashicorp/consul=github.com/hashicorp/consul@v1.5.1" \
		--replace "github.com/gocql/gocql=github.com/grafana/gocql@v0.0.0-20200605141915-ba5dc39ece85" \
		--replace "github.com/prometheus/prometheus=github.com/prometheus/prometheus@v0.42.0" \
    --with "github.com/grafana/xk6-loki=$(PWD)" \
    --with github.com/LeonAdato/xk6-output-statsd@v0.1.1 \
    --output /k6

EOF

RUN cd /home/k6/xk6-loki && make k6

# If you want always using latest version in the master branch of xk6-loki repository.
# But breaking changes may not detected and the test script may outdated.
#RUN xk6 build \
#    --with github.com/grafana/xk6-loki@latest \
#    --with github.com/LeonAdato/xk6-output-statsd@v0.1.1 \
#    --output /k6

# Use alpine like official image https://github.com/grafana/k6/blob/v0.47.0/Dockerfile
FROM alpine:3.18

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 12345 -g 12345 k6

COPY --from=builder /k6 /usr/bin/k6

USER k6
WORKDIR /home/k6
ENTRYPOINT ["k6"]
