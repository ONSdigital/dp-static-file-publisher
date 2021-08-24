package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/aws/aws-sdk-go/aws/session"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthcheck.go -pkg mock . HealthChecker

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetVault(cfg *config.Config) (event.VaultClient, error)
	DoGetImageAPIClient(cfg *config.Config) event.ImageAPIClient
	DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (KafkaConsumer, error)
	DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error)
	DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader
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
	Channels() *dpkafka.ConsumerGroupChannels
}
