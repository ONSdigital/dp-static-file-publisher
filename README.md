dp-static-file-publisher
================
Static file publisher

### Getting started

* Run `make debug`

* You can use the provided kafka producer to send kafka messages that will trigger a file publishing event, for testing proposes.
You can use it by running `go run cmd/producer/main.go`
Then you will need to introduce the source and destination paths, and the message will be sent.

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable         | Default                  | Description
| ---------------------------- | ------------------------ | -----------
| BIND_ADDR                    | :24900                   | The host and port to bind to
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                      | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| VAULT_TOKEN                  | -                        | Vault token required for the client to talk to vault. (Use `make debug` to create a vault token)
| VAULT_ADDR                   | -                        | The vault address
| VAULT_RETRIES                | 3                        | Number of times that a connection to vault will be retried if it fails
| IMAGE_API_URL                | http://localhost:24700   | The image api url
| KAFKA_ADDR                   | localhost:9092           | The list of kafka broker hosts
| STATIC_FILE_PUBLISHED_TOPIC  | static-file-published    | The kafka topic that will be consumed by this service and will trigger a file publishing event
| CONSUMER_GROUP               | dp-static-file-publisher | The kafka consumer-group to consume static-file-published messages

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

