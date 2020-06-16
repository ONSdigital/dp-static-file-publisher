package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	vault "github.com/ONSdigital/dp-vault"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck            bool
	KafkaConsumerPublished bool
	Init                   Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Init: initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// GetVault creates a new vault client
func (e *ExternalServiceList) GetVault(cfg *config.Config) (VaultClient, error) {
	return e.Init.DoGetVault(cfg.VaultToken, cfg.VaultAddress, 3)
}

// GetImageAPIClient creates a new image API client
func (e *ExternalServiceList) GetImageAPIClient(cfg *config.Config) ImageAPIClient {
	return e.Init.DoGetImageAPIClient(cfg.ImageAPIURL)
}

// GetKafkaConsumer returns a kafka consumer group
func (e *ExternalServiceList) GetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafkaConsumer kafka.IConsumerGroup, err error) {
	kafkaConsumer, err = e.Init.DoGetKafkaConsumer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaConsumerPublished = true
	return kafkaConsumer, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// DoGetVault creates a new vault client using dp-vault library
func (e *Init) DoGetVault(vaultToken, vaultAddress string, retries int) (VaultClient, error) {
	return vault.CreateClient(vaultToken, vaultAddress, retries)
}

// DoGetImageAPIClient creates a new image api client using dp-api-clients-go library
func (e *Init) DoGetImageAPIClient(imageAPIURL string) ImageAPIClient {
	return image.NewAPIClient(imageAPIURL)
}

// DoGetKafkaConsumer creates a new Kafka  Consumer Group using dp-kafka library
func (e *Init) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	cgChannels := kafka.CreateConsumerGroupChannels(true)
	return kafka.NewConsumerGroup(ctx, cfg.KafkaAddr, cfg.StaticFilePublishedTopic, cfg.ConsumerGroup, kafka.OffsetNewest, true, cgChannels)
}
