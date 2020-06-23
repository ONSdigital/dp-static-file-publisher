package service

import (
	"context"

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-static-file-publisher/api"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Image API
type Service struct {
	Config        *config.Config
	Server        HTTPServer
	Router        *mux.Router
	API           *api.API
	ServiceList   *ExternalServiceList
	HealthCheck   HealthChecker
	VaultCli      VaultClient
	ImageAPICli   ImageAPIClient
	KafkaConsumer kafka.IConsumerGroup
	S3Public      S3Client
	S3Private     S3Client
}

// Run the service
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (svc *Service, err error) {
	log.Event(ctx, "got service configuration", log.Data{"config": cfg}, log.INFO)

	svc = &Service{
		Config: cfg,
		// HealthCheck: hc,
		ServiceList: serviceList,
	}

	// Get HTTP Router, Server and API
	svc.Router = mux.NewRouter()
	svc.Server = serviceList.GetHTTPServer(cfg.BindAddr, svc.Router)
	svc.API = api.Setup(ctx, svc.Router)

	// Get Vault Client
	svc.VaultCli, err = serviceList.GetVault(cfg)
	if err != nil {
		log.Event(ctx, "could not instantiate vault client", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get Image API Client
	svc.ImageAPICli = serviceList.GetImageAPIClient(cfg)

	// Get Kafka Consumer
	svc.KafkaConsumer, err = serviceList.GetKafkaConsumer(ctx, cfg)
	if err != nil {
		log.Event(ctx, "could not instantiate kafka consumer", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get S3 Clients
	svc.S3Public, svc.S3Private, err = serviceList.GetS3Clients(cfg)
	if err != nil {
		log.Event(ctx, "could not instantiate S3 clients", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get HealthCheck
	svc.HealthCheck, err = serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return nil, err
	}
	if err := svc.registerCheckers(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	svc.Router.StrictSlash(true).Path("/health").HandlerFunc(svc.HealthCheck.Handler)
	svc.HealthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return svc, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutown gracefully closes up
	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownError bool

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.Server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}

		// Close Kafka Consumer, if present (previous stop listening not required)
		if svc.ServiceList.KafkaConsumerPublished {
			if err := svc.KafkaConsumer.Close(ctx); err != nil {
				log.Event(ctx, "failed to shutdown kafka consumer group", log.Error(err), log.ERROR)
				hasShutdownError = true
			}
		}

		// close API
		if err := svc.API.Close(ctx); err != nil {
			log.Event(ctx, "error closing API", log.Error(err), log.ERROR)
			hasShutdownError = true
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds all the necessary checkers to healthcheck. Please, only call this function after all dependencies are instanciated
func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	if err = svc.HealthCheck.AddCheck("Vault", svc.VaultCli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add vault client checker", log.ERROR, log.Error(err))
	}

	if err = svc.HealthCheck.AddCheck("Image API", svc.ImageAPICli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add image api client checker", log.ERROR, log.Error(err))
	}

	if err = svc.HealthCheck.AddCheck("Kafka Consumer", svc.KafkaConsumer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add kafka consumer checker", log.ERROR, log.Error(err))
	}

	if err = svc.HealthCheck.AddCheck("S3 Public", svc.S3Public.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add public s3 client checker", log.ERROR, log.Error(err))
	}

	if err = svc.HealthCheck.AddCheck("S3 Private", svc.S3Private.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add private s3 client checker", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
