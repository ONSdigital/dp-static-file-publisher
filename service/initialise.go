package service

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpkafka "github.com/ONSdigital/dp-kafka"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
	dps3 "github.com/ONSdigital/dp-s3"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	dpvault "github.com/ONSdigital/dp-vault"
	"github.com/aws/aws-sdk-go/aws/session"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	Vault                  bool
	ImageAPI               bool
	HealthCheck            bool
	KafkaConsumerPublished bool
	S3Private              bool
	S3Public               bool
	Init                   Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Vault:                  false,
		ImageAPI:               false,
		HealthCheck:            false,
		KafkaConsumerPublished: false,
		S3Private:              false,
		S3Public:               false,
		Init:                   initialiser,
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

// GetVault creates a Vault client and sets the Vault flag to true
func (e *ExternalServiceList) GetVault(cfg *config.Config) (event.VaultClient, error) {
	vault, err := e.Init.DoGetVault(cfg)
	if err != nil {
		return nil, err
	}
	e.Vault = true
	return vault, nil
}

// GetImageAPIClient creates an ImageAPI client and sets the ImageAPI flag to true
func (e *ExternalServiceList) GetImageAPIClient(cfg *config.Config) event.ImageAPIClient {
	imageAPI := e.Init.DoGetImageAPIClient(cfg)
	e.ImageAPI = true
	return imageAPI
}

// GetKafkaConsumer creates a Kafka consumer and sets the consumer flag to true
func (e *ExternalServiceList) GetKafkaConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error) {
	kafkaConsumer, err := e.Init.DoGetKafkaConsumer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaConsumerPublished = true
	return kafkaConsumer, nil
}

// GetS3Clients returns S3 clients private and public. They share the same AWS session.
func (e *ExternalServiceList) GetS3Clients(cfg *config.Config) (s3Private event.S3Reader, s3Public event.S3Writer, err error) {
	s3Public, err = e.Init.DoGetS3Client(cfg.AwsRegion, cfg.PublicBucketName, false)
	if err != nil {
		return nil, nil, err
	}
	e.S3Public = true
	s3Private = e.Init.DoGetS3ClientWithSession(cfg.PrivateBucketName, !cfg.EncryptionDisabled, s3Public.Session())
	e.S3Private = true
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

// DoGetVault returns a VaultClient if encryption is enabled
func (e *Init) DoGetVault(cfg *config.Config) (event.VaultClient, error) {
	if cfg.EncryptionDisabled {
		return nil, nil
	}
	vault, err := dpvault.CreateClient(cfg.VaultToken, cfg.VaultAddress, 3)
	if err != nil {
		return nil, err
	}
	return vault, nil
}

// DoGetImageAPIClient returns an Image API client
func (e *Init) DoGetImageAPIClient(cfg *config.Config) event.ImageAPIClient {
	return image.NewAPIClient(cfg.ImageAPIURL)
}

// DoGetKafkaConsumer returns a Kafka Consumer group
func (e *Init) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error) {
	cgChannels := dpkafka.CreateConsumerGroupChannels(true)
	return dpkafka.NewConsumerGroup(
		ctx,
		cfg.KafkaAddr,
		cfg.StaticFilePublishedTopic,
		cfg.ConsumerGroup,
		dpkafka.OffsetNewest,
		true,
		cgChannels,
	)
}

// DoGetS3Client creates a new S3Client for the provided AWS region and bucket name.
func (e *Init) DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
	return dps3.NewUploader(awsRegion, bucketName, encryptionEnabled)
}

// DoGetS3ClientWithSession creates a new S3Clienter (extension of S3Client with Upload operations)
// for the provided bucket name, using an existing AWS session
func (e *Init) DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
	return dps3.NewClientWithSession(bucketName, encryptionEnabled, s)
}
