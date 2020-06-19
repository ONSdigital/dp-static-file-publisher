package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-static-file-publisher/config"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthcheck.go -pkg mock . HealthChecker
//go:generate moq -out mock/vault.go -pkg mock . VaultClient
//go:generate moq -out mock/image.go -pkg mock . ImageAPIClient

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetVault(vaultToken, vaultAddress string, retries int) (VaultClient, error)
	DoGetImageAPIClient(imageAPIURL string) ImageAPIClient
	DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error)
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

// VaultClient defines the required methods from dp-vault client
type VaultClient interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// ImageAPIClient defines the required methods from dp-api-clients-go ImageAPI
type ImageAPIClient interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
