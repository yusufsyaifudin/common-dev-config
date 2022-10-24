# Benthos YAML Sample

This repository contains bunch of working example of using 
[Benthos](https://www.benthos.dev/) - a fancy stream processing made operationally mundane.


* [x] [kafka-to-http](/benthos/kafka-to-http) Will consume message from Kafka topic and post it into HTTP Service.
  This useful when you have a system that need to call another system and not require its responses.
  For example, you create Sign-Up service that once user register, you push the data to Kafka and then use Benthos
  to consume the data and post it to 3rd party service such as [SendGrid](https://docs.sendgrid.com/api-reference/mail-send/mail-send) or your homegrown service.
  
* [ ] [add yours](#) Add more sample