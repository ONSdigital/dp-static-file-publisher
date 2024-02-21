# DP Static File Publisher

## Introduction
The Static File Publisher API is part of the [Static Files System](https://github.com/ONSdigital/dp-static-files-compose).

It receives messages on two Kafka topics `static-file-published` (which has been deprecated, but not yet removed) and
`static-file-published-v2` which is the new topic that the [Files API](https://github.com/ONSdigital/dp-files-api) sends
publication messages on.

## Getting started

* Run `make debug`

* You can use the provided kafka producer to send kafka messages that will trigger a file publishing event, for testing proposes.
You can use it by running `go run cmd/producer/main.go`
Then you will need to introduce the source and destination paths, and the message will be sent.

## Dependencies

* No further dependencies other than those defined in `go.mod`

## Configuration

| Environment variable           | Default                  | Description                                                                                                        |
|--------------------------------|--------------------------|--------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                      | :24900                   | The host and port to bind to                                                                                       |
| GRACEFUL_SHUTDOWN_TIMEOUT      | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_INTERVAL           | 30s                      | Time between self-healthchecks (`time.Duration` format)                                                            |
| HEALTHCHECK_CRITICAL_TIMEOUT   | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| FILES_API_URL                  |                          | The URL of the dp-files-api                                                                                        |
| IMAGE_API_URL                  | http://localhost:24700   | The image api url                                                                                                  |
| KAFKA_ADDR                     | localhost:9092           | The list of kafka broker hosts                                                                                     |
| KAFKA_VERSION                  | `1.0.2`                  | The version of Kafka                                                                                               |
| KAFKA_CONSUMER_WORKERS         | 1                        | The maximum number of parallel kafka consumers                                                                     |
| KAFKA_SEC_PROTO                | _unset_   (only `TLS`)   | if set to `TLS`, kafka connections will use TLS                                                                    |
| KAFKA_SEC_CLIENT_KEY           | _unset_                  | PEM [2] for the client key (optional, used for client auth) [1]                                                    |
| KAFKA_SEC_CLIENT_CERT          | _unset_                  | PEM [2] for the client certificate (optional, used for client auth) [1]                                            |
| KAFKA_SEC_CA_CERTS             | _unset_                  | PEM [2] of CA cert chain if using private CA for the server cert [1]                                               |
| KAFKA_SEC_SKIP_VERIFY          | false                    | ignore server certificate issues if set to `true` [1]                                                              |
| CONSUMER_GROUP                 | dp-static-file-publisher | The kafka consumer-group to consume static-file-published messages                                                 |
| STATIC_FILE_PUBLISHED_TOPIC    | static-file-published    | The kafka topic that will be consumed by this service and will trigger a file publishing event [DEPRECATED]        |
| STATIC_FILE_PUBLISHED_TOPIC_V2 | static-file-published-v2 | The kafka topic that will be consumed by this service and will trigger a file publishing event from dp-files-api   |
| S3_LOCAL_URL                   |                          | S3 Configuration for integration tests                                                                             |
| S3_LOCAL_ID                    |                          | S3 Configuration for integration tests                                                                             |
| S3_LOCAL_SECRET                |                          | S3 Configuration for integration tests                                                                             |

**Notes:**

1. For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

