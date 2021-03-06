package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-static-file-publisher
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"             json:"-"`
	EncryptionDisabled         bool          `envconfig:"ENCRYPTION_DISABLED"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	VaultToken                 string        `envconfig:"VAULT_TOKEN"                   json:"-"`
	VaultAddress               string        `envconfig:"VAULT_ADDR"`
	VaultRetries               int           `envconfig:"VAULT_RETRIES"`
	VaultPath                  string        `envconfig:"VAULT_PATH"`
	ImageAPIURL                string        `envconfig:"IMAGE_API_URL"`
	KafkaAddr                  []string      `envconfig:"KAFKA_ADDR"`
	StaticFilePublishedTopic   string        `envconfig:"STATIC_FILE_PUBLISHED_TOPIC"`
	ConsumerGroup              string        `envconfig:"CONSUMER_GROUP"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	PrivateBucketName          string        `envconfig:"S3_PRIVATE_BUCKET_NAME"`
	PublicBucketName           string        `envconfig:"S3_PUBLIC_BUCKET_NAME"`
	PublicBucketURL            string        `envconfig:"S3_PUBLIC_BUCKET_URL"`
}

var cfg *Config

// Get returns the default config with any modifications through environment variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                   "localhost:24900",
		ServiceAuthToken:           "4424A9F2-B903-40F4-85F1-240107D1AFAF",
		EncryptionDisabled:         false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		VaultToken:                 "",
		VaultAddress:               "",
		VaultRetries:               3,
		VaultPath:                  "secret/shared/psk",
		ImageAPIURL:                "http://localhost:24700",
		KafkaAddr:                  []string{"localhost:9092"},
		StaticFilePublishedTopic:   "static-file-published",
		ConsumerGroup:              "dp-static-file-publisher",
		AwsRegion:                  "eu-west-1",
		PrivateBucketName:          "csv-exported",
		PublicBucketName:           "static-develop",
		PublicBucketURL:            "https://static-develop.s3.eu-west-1.amazonaws.com",
	}

	return cfg, envconfig.Process("", cfg)
}
