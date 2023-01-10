package service

import (
	"context"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the Image API
type Service struct {
	Config                      *config.Config
	Server                      HTTPServer
	Router                      *mux.Router
	ServiceList                 *ExternalServiceList
	HealthCheck                 HealthChecker
	ImageAPICli                 event.ImageAPIClient
	KafkaImagePublishedConsumer KafkaConsumer
	KafkaFilePublishedConsumer  KafkaConsumer
	S3Public                    event.S3Writer
	S3Private                   event.S3Reader
	VaultCli                    event.VaultClient
	FilesClient                 file.FilesService
}

// Run the service
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (svc *Service, err error) {
	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

	svc = &Service{
		Config: cfg,
		// HealthCheck: hc,
		ServiceList: serviceList,
	}

	// Get HTTP Router, Server and API
	svc.Router = mux.NewRouter()
	svc.Server = serviceList.GetHTTPServer(cfg.BindAddr, svc.Router)

	// Get Vault Client
	svc.VaultCli, err = serviceList.GetVault(cfg)
	if err != nil {
		log.Fatal(ctx, "could not instantiate vault client", err)
		return nil, err
	}

	// Get Image API Client
	svc.ImageAPICli = serviceList.GetImageAPIClient(cfg)

	// Get S3 Clients
	svc.S3Private, svc.S3Public, err = serviceList.GetS3Clients(cfg)
	if err != nil {
		log.Fatal(ctx, "could not instantiate S3 clients", err)
		return nil, err
	}

	// Initialise KafkaImagePublishedConsumer
	svc.KafkaImagePublishedConsumer, err = serviceList.GetKafkaImagePublishedConsumer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "could not instantiate KafkaImagePublishedConsumer", err)
		return nil, err
	}
	handler := &event.ImagePublishedHandler{
		AuthToken:       cfg.ServiceAuthToken,
		S3Private:       svc.S3Private,
		S3Public:        svc.S3Public,
		VaultCli:        svc.VaultCli,
		VaultPath:       cfg.VaultPath,
		ImageAPICli:     svc.ImageAPICli,
		PublicBucketURL: cfg.PublicBucketURL,
	}
	if err = svc.KafkaImagePublishedConsumer.RegisterBatchHandler(ctx, handler.KafkaHandler); err != nil {
		log.Fatal(ctx, "failed to register image published message handler", err)
		return nil, err
	}
	if err = svc.KafkaImagePublishedConsumer.Start(); err != nil {
		log.Fatal(ctx, "Could not start KafkaImagePublishedConsumer", err)
		return nil, err
	}

	// Initialise KafkaFilePublishedConsumer
	svc.KafkaFilePublishedConsumer, err = serviceList.GetKafkaFilePublishedConsumer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "Could not instantiate KafkaFilePublishedConsumer", err)
		return nil, err
	}
	dc, err := getDecrypterCopier(ctx, cfg, serviceList, svc)
	if err != nil {
		return nil, err
	}
	if err = svc.KafkaFilePublishedConsumer.RegisterBatchHandler(ctx, dc.HandleFilePublishMessage); err != nil {
		log.Fatal(ctx, "failed to register file published message handler", err)
		return nil, err
	}
	if err = svc.KafkaFilePublishedConsumer.Start(); err != nil {
		log.Fatal(ctx, "Could not start KafkaFilePublishedConsumer", err)
		return nil, err
	}

	// Get HealthCheck
	svc.HealthCheck, err = serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
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

func getDecrypterCopier(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, svc *Service) (file.DecrypterCopier, error) {
	publicClient, err := serviceList.GetS3ClientV2(cfg, cfg.PublicBucketName)
	if err != nil {
		log.Fatal(ctx, "Could not instantiate public S3 v2 client", err)
		return file.DecrypterCopier{}, err
	}

	privateClient, err := serviceList.GetS3ClientV2(cfg, cfg.PrivateBucketName)
	if err != nil {
		log.Fatal(ctx, "Could not instantiate private S3 v2 client", err)
		return file.DecrypterCopier{}, err
	}

	svc.FilesClient = serviceList.GetFilesService(ctx, cfg)

	return file.NewDecrypterCopier(publicClient, privateClient, svc.VaultCli, cfg.VaultPath, svc.FilesClient), nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
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
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// Request Kafka consumers to stop, if they're present
		if svc.ServiceList.KafkaImagePublishedConsumer {
			if err := svc.KafkaImagePublishedConsumer.Stop(); err != nil {
				log.Error(ctx, "failed to stop KafkaImagePublishedConsumer", err)
				hasShutdownError = true
			}
		}
		if svc.ServiceList.KafkaFilePublishedConsumer {
			if err := svc.KafkaFilePublishedConsumer.Stop(); err != nil {
				log.Error(ctx, "failed to stop KafkaFilePublishedConsumer", err)
				hasShutdownError = true
			}
		}

		// Wait for them to stop
		if svc.ServiceList.KafkaImagePublishedConsumer {
			svc.KafkaImagePublishedConsumer.StateWait(kafka.Stopped)
		}
		if svc.ServiceList.KafkaFilePublishedConsumer {
			svc.KafkaFilePublishedConsumer.StateWait(kafka.Stopped)
		}

		// Close the Kafka consumers
		if svc.ServiceList.KafkaImagePublishedConsumer {
			if err := svc.KafkaImagePublishedConsumer.Close(ctx); err != nil {
				log.Error(ctx, "failed to close KafkaImagePublishedConsumer", err)
				hasShutdownError = true
			}
		}
		if svc.ServiceList.KafkaFilePublishedConsumer {
			if err := svc.KafkaFilePublishedConsumer.Close(ctx); err != nil {
				log.Error(ctx, "failed to close KafkaFilePublishedConsumer", err)
				hasShutdownError = true
			}
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

// registerCheckers adds all the necessary checkers to healthcheck. Please, only call this function after all dependencies are instanciated
func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	// nolint:typecheck // disabled due to failing in CI, but works locally on mac
	if svc.VaultCli != nil {
		if err = svc.HealthCheck.AddCheck("Vault", svc.VaultCli.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "failed to add vault client checker", err)
		}
	}

	if err = svc.HealthCheck.AddCheck("Image API", svc.ImageAPICli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add image api client checker", err)
	}

	if err = svc.HealthCheck.AddCheck("Kafka ImagePublished Consumer", svc.KafkaImagePublishedConsumer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add KafkaImagePublishedConsumer checker", err)
	}

	if err = svc.HealthCheck.AddCheck("Kafka FilePublished Consumer", svc.KafkaFilePublishedConsumer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add KafkaFilePublishedConsumer checker", err)
	}

	if err = svc.HealthCheck.AddCheck("S3 Public", svc.S3Public.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add public s3 client checker", err)
	}

	if err = svc.HealthCheck.AddCheck("S3 Private", svc.S3Private.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add private s3 client checker", err)
	}

	if err = svc.HealthCheck.AddCheck("Files Service", svc.FilesClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add files api client checker", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
