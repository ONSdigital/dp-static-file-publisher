package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	dphttp "github.com/ONSdigital/dp-net/http"
	s3client "github.com/ONSdigital/dp-s3"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/aws/aws-sdk-go/aws/session"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck            bool
	KafkaConsumerPublished bool
	S3Public               bool
	S3Private              bool
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
func (e *ExternalServiceList) GetVault(cfg *config.Config) (event.VaultClient, error) {
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

// GetS3Clients returns S3 clients public and private. They share the same AWS session.
func (e *ExternalServiceList) GetS3Clients(cfg *config.Config) (s3Public event.S3Uploader, s3Private event.S3Client, err error) {
	s3Private, err = e.Init.DoGetS3Client(cfg.AwsRegion, cfg.PrivateBucketName, true)
	if err != nil {
		return nil, nil, err
	}
	e.S3Private = true
	s3Public = e.Init.DoGetS3UploaderWithSession(cfg.PublicBucketName, true, s3Private.Session())
	e.S3Public = true
	return
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
func (e *Init) DoGetVault(vaultToken, vaultAddress string, retries int) (event.VaultClient, error) {
	return vault.CreateClient(vaultToken, vaultAddress, retries)
}

// DoGetImageAPIClient creates a new image api client using dp-api-clients-go library
func (e *Init) DoGetImageAPIClient(imageAPIURL string) ImageAPIClient {
	return image.NewAPIClient(imageAPIURL)
}

// DoGetKafkaConsumer creates a new Kafka Consumer Group using dp-kafka library
func (e *Init) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	cgChannels := kafka.CreateConsumerGroupChannels(true)
	return kafka.NewConsumerGroup(ctx, cfg.KafkaAddr, cfg.StaticFilePublishedTopic, cfg.ConsumerGroup, kafka.OffsetNewest, true, cgChannels)
}

// DoGetS3Client creates a new S3Client for the provided AWS region and bucket name.
func (e *Init) DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Client, error) {
	return s3client.NewClient(awsRegion, bucketName, encryptionEnabled)
}

// DoGetS3UploaderWithSession creates a new S3Uploader (extension of S3Client with Upload operations)
// for the provided bucket name, using an existing AWS session
func (e *Init) DoGetS3UploaderWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Uploader {
	return s3client.NewUploaderWithSession(bucketName, encryptionEnabled, s)
}
