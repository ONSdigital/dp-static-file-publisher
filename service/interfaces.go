package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-static-file-publisher/file"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafkaV2 "github.com/ONSdigital/dp-kafka/v2"
	kafkaV3 "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/aws/aws-sdk-go/aws/session"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthcheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/kafka.go -pkg mock . KafkaConsumerV3

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetVault(cfg *config.Config) (event.VaultClient, error)
	DoGetImageAPIClient(cfg *config.Config) event.ImageAPIClient
	DoGetKafkaImagePublishedConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumerV3, error)
	DoGetKafkaFilePublishedConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumerV3, error)
	DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error)
	DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader
	DoGetS3ClientV2(awsRegion, bucketName string) (file.S3ClientV2, error)
	DoGetFilesService(ctx context.Context, cfg *config.Config) file.FilesService
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

type KafkaConsumer interface {
	Initialise(ctx context.Context) error
	IsInitialised() bool
	StopListeningToConsumer(ctx context.Context) (err error)
	Close(ctx context.Context) (err error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Channels() *kafkaV2.ConsumerGroupChannels
}

type KafkaConsumerV3 interface {
	Start() error
	RegisterBatchHandler(ctx context.Context, batchHandler kafkaV3.BatchHandler) error
	Stop() error
	StateWait(state kafkaV3.State)
	Checker(ctx context.Context, state *healthcheck.CheckState) error

	Channels() *kafkaV3.ConsumerGroupChannels
	Close(ctx context.Context) (err error)
	Initialise(ctx context.Context) error
	IsInitialised() bool
	State() kafkaV3.State
	RegisterHandler(ctx context.Context, h kafkaV3.Handler) error
	OnHealthUpdate(status string)
	StopAndWait() error
	LogErrors(ctx context.Context)
}
