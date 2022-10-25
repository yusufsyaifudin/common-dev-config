# Benthos YAML Sample

This repository contains bunch of working example of using 
[Benthos](https://www.benthos.dev/) - a fancy stream processing made operationally mundane.


* [x] [echo-server](/benthos/echo-server) Will create a service with URL http://127.0.0.1/post where you can POST the request body.
  Then it will return HTTP Status 200 OK with modified response body JSON containing request header, request body, etc.
  This is similar like [https://postman-echo.com/post](https://postman-echo.com/post) in [here.](https://documenter.getpostman.com/view/5025623/SWTG5aqV)

* [x] [kafka-batch-consumer](/benthos/kafka-batch-consumer) Consume the message payload from Kafka topic and send to output as batch.
  In means that Benthos will collect and wait for several seconds before passing the message to the output sink.
  This is useful when we want to add write performance to output, because when the Kafka traffic is large, Benthos will send it as a batch.

* [x] [kafka-to-es](/benthos/kafka-to-es) Will consume message from Kafka and persist it in Elasticsearch without message payload modification.
  This is useful when we want to create an Audit Logs system.

* [x] [kafka-to-http](/benthos/kafka-to-http) Will consume message from Kafka topic and post it into HTTP Service.
  This useful when you have a system that need to call another system and not require its responses.
  For example, you create Sign-Up service that once user register, you push the data to Kafka and then use Benthos
  to consume the data and post it to 3rd party service such as [SendGrid](https://docs.sendgrid.com/api-reference/mail-send/mail-send) or your homegrown service.

* [ ] [add yours](#) Add more sample

## How to run

This is common way to run the example using Docker Compose.

1. `cd` to the path of the Benthos stream that you want to run
2. `cp .env.sample .env` and modify the value based on your need: i.e change the target IP, etc
3. Ensure that the required service is up and unreachable (if you behind VPN, ensure that it can be accessed).
4. Run `docker compose up` and it should read the `.env` that we create.
5. _optional_ use some Bash script included in each folder to produce the Kafka payload, for example `publish_message.sh`.
6. Read the README.md in each folder for detailed step and information.

Alternatively, you can run using `run.sh` that exists in each Benthos stream folder.
