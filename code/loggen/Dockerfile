# Build the k6 binary with the extension
FROM golang:1.21-alpine3.18 as builder

RUN apk update && apk add bash git

WORKDIR /src
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN env CGO_ENABLED=0 go build -o log-generator cli/main.go
RUN env CGO_ENABLED=0 go build -o log-generator-server server/main.go

FROM alpine:3.18

 # install coreutils for precission in time
RUN apk update && apk add bash coreutils

WORKDIR /app
COPY --from=builder /src/log-generator /app/log-generator
COPY --from=builder /src/log-generator-server /app/log-generator-server

ENTRYPOINT ["/app/log-generator"]
