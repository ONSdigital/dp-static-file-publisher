package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-static-file-publisher
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"            json:"-"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	ImageAPIURL                string        `envconfig:"IMAGE_API_URL"`
	KafkaAddr                  []string      `envconfig:"KAFKA_ADDR"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	KafkaSecProtocol           string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts            string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert         string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey          string        `envconfig:"KAFKA_SEC_CLIENT_KEY"          json:"-"`
	KafkaSecSkipVerify         bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	KafkaConsumerWorkers       int           `envconfig:"KAFKA_CONSUMER_WORKERS"`
	KafkaMinimumHealthyBrokers int           `envconfig:"KAFKA_MIN_HEALTHY_BROKERS"`
	KafkaBatchSize             int           `envconfig:"KAFKA_BATCH_SIZE"`
	KafkaBatchWaitTime         time.Duration `envconfig:"KAFKA_BATCH_WAIT_TIME"`
	ImageFilePublishedTopic    string        `envconfig:"STATIC_FILE_PUBLISHED_TOPIC"`
	StaticFilePublishedTopic   string        `envconfig:"STATIC_FILE_PUBLISHED_TOPIC_V2"`
	ConsumerGroup              string        `envconfig:"CONSUMER_GROUP"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	PrivateBucketName          string        `envconfig:"S3_PRIVATE_BUCKET_NAME"`
	PublicBucketName           string        `envconfig:"S3_PUBLIC_BUCKET_NAME"`
	PublicBucketURL            string        `envconfig:"S3_PUBLIC_BUCKET_URL"`
	LocalS3URL                 string        `envconfig:"S3_LOCAL_URL"`
	LocalS3ID                  string        `envconfig:"S3_LOCAL_ID"`
	LocalS3Secret              string        `envconfig:"S3_LOCAL_SECRET"`
	FilesAPIURL                string        `envconfig:"FILES_API_URL"`
	KafkaConfig                KafkaConfig
}

type KafkaConfig struct {
	Addr                      []string `envconfig:"KAFKA_ADDR"                            json:"-"`
	ConsumerMinBrokersHealthy int      `envconfig:"KAFKA_CONSUMER_MIN_BROKERS_HEALTHY"`
	ProducerMinBrokersHealthy int      `envconfig:"KAFKA_PRODUCER_MIN_BROKERS_HEALTHY"`
	Version                   string   `envconfig:"KAFKA_VERSION"`
	OffsetOldest              bool     `envconfig:"KAFKA_OFFSET_OLDEST"`
	NumWorkers                int      `envconfig:"KAFKA_NUM_WORKERS"`
	MaxBytes                  int      `envconfig:"KAFKA_MAX_BYTES"`
	SecProtocol               string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts                string   `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecClientKey              string   `envconfig:"KAFKA_SEC_CLIENT_KEY"                  json:"-"`
	SecClientCert             string   `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecSkipVerify             bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	ImageFilePublishedTopic   string   `envconfig:"STATIC_FILE_PUBLISHED_TOPIC"`
	StaticFilePublishedTopic  string   `envconfig:"STATIC_FILE_PUBLISHED_TOPIC_V2"`
}

var cfg *Config

// Get returns the default config with any modifications through environment variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                   ":24900",
		ServiceAuthToken:           "4424A9F2-B903-40F4-85F1-240107D1AFAF",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		ImageAPIURL:                "http://localhost:24700",
		KafkaAddr:                  []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		KafkaVersion:               "1.0.2",
		KafkaConsumerWorkers:       1,
		KafkaMinimumHealthyBrokers: 1,
		KafkaBatchSize:             500,
		KafkaBatchWaitTime:         50 * time.Millisecond,
		ImageFilePublishedTopic:    "static-file-published",
		StaticFilePublishedTopic:   "static-file-published-v2",
		ConsumerGroup:              "dp-static-file-publisher",
		AwsRegion:                  "eu-west-2",
		PrivateBucketName:          "csv-exported",
		PublicBucketName:           "static-develop",
		PublicBucketURL:            "https://static-develop.s3.eu-west-2.amazonaws.com",
		FilesAPIURL:                "http://localhost:26900",
	}

	return cfg, envconfig.Process("", cfg)
}
