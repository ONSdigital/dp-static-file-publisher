package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-static-file-publisher
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	VaultToken                 string        `envconfig:"VAULT_TOKEN"                   json:"-"`
	VaultAddress               string        `envconfig:"VAULT_ADDR"`
	VaultRetries               int           `envconfig:"VAULT_RETRIES"`
	ImageAPIURL                string        `envconfig:"IMAGE_API_URL"`
	KafkaAddr                  []string      `envconfig:"KAFKA_ADDR"`
	StaticFilePublishedTopic   string        `envconfig:"STATIC_FILE_PUBLISHED_TOPIC"`
	ConsumerGroup              string        `envconfig:"CONSUMER_GROUP"`
}

var cfg *Config

// Get returns the default config with any modifications through environment variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                   ":24900",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		VaultToken:                 "",
		VaultAddress:               "",
		VaultRetries:               3,
		ImageAPIURL:                "http://localhost:24700",
		KafkaAddr:                  []string{"localhost:9092"},
		StaticFilePublishedTopic:   "static-file-published",
		ConsumerGroup:              "dp-static-file-publisher",
	}

	return cfg, envconfig.Process("", cfg)
}
