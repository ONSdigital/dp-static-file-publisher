package service

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/ONSdigital/dp-api-clients-go/image"
	files "github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	dphttp "github.com/ONSdigital/dp-net/http"
	dps3 "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	ImageAPI                    bool
	FilesService                bool
	HealthCheck                 bool
	KafkaImagePublishedConsumer bool
	KafkaFilePublishedConsumer  bool
	S3Private                   bool
	S3Public                    bool
	S3Client                    bool
	Init                        Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		ImageAPI:                    false,
		HealthCheck:                 false,
		KafkaImagePublishedConsumer: false,
		KafkaFilePublishedConsumer:  false,
		S3Private:                   false,
		S3Public:                    false,
		S3Client:                    false,
		FilesService:                false,
		Init:                        initialiser,
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

// GetImageAPIClient creates an ImageAPI client and sets the ImageAPI flag to true
func (e *ExternalServiceList) GetImageAPIClient(cfg *config.Config) event.ImageAPIClient {
	imageAPI := e.Init.DoGetImageAPIClient(cfg)
	e.ImageAPI = true
	return imageAPI
}

// GetFilesService creates files service  and sets the FilesService flag to true
func (e *ExternalServiceList) GetFilesService(ctx context.Context, cfg *config.Config) file.FilesService {
	client := e.Init.DoGetFilesService(ctx, cfg)
	e.FilesService = true
	return client
}

// GetKafkaImagePublishedConsumer creates a Kafka consumer and sets the consumer flag to true
func (e *ExternalServiceList) GetKafkaImagePublishedConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error) {
	imagePublishedConsumer, err := e.Init.DoGetKafkaImagePublishedConsumer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaImagePublishedConsumer = true
	return imagePublishedConsumer, nil
}

// GetKafkaFilePublishedConsumer creates a Kafka consumer and sets the consumer flag to true
func (e *ExternalServiceList) GetKafkaFilePublishedConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error) {
	filePublishedConsumer, err := e.Init.DoGetKafkaFilePublishedConsumer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaFilePublishedConsumer = true
	return filePublishedConsumer, nil
}

// GetS3Clients returns S3 clients private and public. They share the same AWS session.
func (e *ExternalServiceList) GetS3Clients(ctx context.Context, cfg *config.Config) (s3Private, s3Public file.S3Client, err error) {
	s3Public, err = e.Init.DoGetS3Client(ctx, cfg.AwsRegion, cfg.PublicBucketName)
	if err != nil {
		return nil, nil, err
	}
	e.S3Public = true
	s3Private, err = e.Init.DoGetS3ClientWithConfig(cfg.PrivateBucketName, s3Public.Config())
	if err != nil {
		return nil, nil, err
	}
	e.S3Private = true
	return
}

// GetS3Client returns S3 clients private and public. They share the same AWS session.
func (e *ExternalServiceList) GetS3Client(ctx context.Context, cfg *config.Config, bucketName string) (file.S3Client, error) {
	s3Client, err := e.Init.DoGetS3Client(ctx, cfg.AwsRegion, bucketName)
	if err != nil {
		return nil, err
	}
	e.S3Client = true
	return s3Client, nil
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

// DoGetImageAPIClient returns an Image API client
func (e *Init) DoGetImageAPIClient(cfg *config.Config) event.ImageAPIClient {
	return image.NewAPIClient(cfg.ImageAPIURL)
}

// DoGetFilesService returns a files service backend
func (e *Init) DoGetFilesService(_ context.Context, cfg *config.Config) file.FilesService {
	apiClient := files.NewAPIClient(cfg.FilesAPIURL, cfg.ServiceAuthToken)
	return apiClient
}

// DoGetKafkaImagePublishedConsumer returns a Kafka Consumer group
func (e *Init) DoGetKafkaImagePublishedConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error) {
	return e.DoGetKafkaTopicConsumer(ctx, cfg, cfg.ImageFilePublishedTopic)
}

func (e *Init) DoGetKafkaTopicConsumer(ctx context.Context, cfg *config.Config, topic string) (KafkaConsumer, error) {
	kafkaOffset := kafka.OffsetOldest

	gc := kafka.ConsumerGroupConfig{
		KafkaVersion:      &cfg.KafkaVersion,
		Offset:            &kafkaOffset,
		MinBrokersHealthy: &cfg.KafkaMinimumHealthyBrokers,
		Topic:             topic,
		GroupName:         cfg.ConsumerGroup,
		BrokerAddrs:       cfg.KafkaAddr,
		NumWorkers:        &cfg.KafkaConsumerWorkers,
		BatchSize:         &cfg.KafkaBatchSize,
		BatchWaitTime:     &cfg.KafkaBatchWaitTime,
	}

	if cfg.KafkaSecProtocol == "TLS" {
		gc.SecurityConfig = &kafka.SecurityConfig{
			RootCACerts:        cfg.KafkaSecCACerts,
			ClientCert:         cfg.KafkaSecClientCert,
			ClientKey:          cfg.KafkaSecClientKey,
			InsecureSkipVerify: cfg.KafkaSecSkipVerify,
		}
	}

	return kafka.NewConsumerGroup(ctx, &gc)
}

func (e *Init) DoGetKafkaFilePublishedConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error) {
	return e.DoGetKafkaTopicConsumer(ctx, cfg, cfg.StaticFilePublishedTopic)
}

func (e *Init) DoGetS3Client(ctx context.Context, awsRegion, bucketName string) (file.S3Client, error) {
	var err error
	cfg, _ := config.Get()

	// If running locally using localstack, `Endpoint` needs to be defined and `S3ForcePathStyle`
	// needs to be set to `true`. This ensures the client makes requests to path style urls rather
	// than virtual hosted style see https://docs.localstack.cloud/user-guide/aws/s3/#path-style-and-virtual-hosted-style-requests
	var AWSConfig aws.Config
	var client *dps3.Client
	if cfg.LocalS3URL != "" {
		AWSConfig, err = awsConfig.LoadDefaultConfig(
			ctx,
			awsConfig.WithRegion(awsRegion),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.LocalS3ID, cfg.LocalS3Secret, "")),
		)

		if err != nil {
			return nil, err
		}

		client = dps3.NewClientWithConfig(bucketName, AWSConfig, func(options *s3.Options) {
			options.BaseEndpoint = aws.String(cfg.LocalS3URL)
			options.UsePathStyle = true
		})

	} else {
		AWSConfig, err = awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(awsRegion))

		if err != nil {
			return nil, err
		}

		client = dps3.NewClientWithConfig(bucketName, AWSConfig)
	}

	return client, nil
}

// DoGetS3ClientWithSession creates a new S3Client (extension of S3Client with Upload operations)
// for the provided bucket name, using an existing AWS session
func (e *Init) DoGetS3ClientWithConfig(bucketName string, cfg aws.Config) (file.S3Client, error) {
	return dps3.NewClientWithConfig(bucketName, cfg), nil
}
