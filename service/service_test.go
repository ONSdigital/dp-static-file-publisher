package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-kafka/kafkatest"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"github.com/ONSdigital/dp-static-file-publisher/service/mock"
	serviceMock "github.com/ONSdigital/dp-static-file-publisher/service/mock"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx              = context.Background()
	testBuildTime    = "BuildTime"
	testGitCommit    = "GitCommit"
	testVersion      = "Version"
	errVault         = errors.New("Vault client error")
	errKafkaConsumer = errors.New("Kafka consumer error")
	errS3            = errors.New("S3 session error")
	errServer        = errors.New("HTTP Server error")
	errHealthcheck   = errors.New("healthCheck error")
)

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetVaultErr = func(vaultToken string, vaultAddress string, retries int) (service.VaultClient, error) {
	return nil, errVault
}

var funcDoGetKafkaConsumerErr = func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	return nil, errKafkaConsumer
}

var funcDoGetS3ClientFuncErr = func(awsRegion string, bucketName string, encryptionEnabled bool, s *session.Session) (service.S3Client, error) {
	return nil, errS3
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
	return nil
}

func TestRun(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		vaultMock := &serviceMock.VaultClientMock{}

		imageAPIClientMock := &serviceMock.ImageAPIClientMock{}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		kafkaConsumerMock := &kafkatest.IConsumerGroupMock{}

		s3Session := &session.Session{}
		s3Mock := &serviceMock.S3ClientMock{
			SessionFunc: func() *session.Session { return s3Session },
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		failingServerMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return errServer
			},
		}

		funcDoGetVaultOK := func(vaultToken string, vaultAddress string, retries int) (service.VaultClient, error) {
			return vaultMock, nil
		}

		funcDoGetImageAPIClientFuncOK := func(imageAPIURL string) service.ImageAPIClient {
			return imageAPIClientMock
		}

		funcDoGetKafkaConsumerOK := func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
			return kafkaConsumerMock, nil
		}

		funcDoGetS3ClientFuncOK := func(awsRegion string, bucketName string, encryptionEnabled bool, s *session.Session) (service.S3Client, error) {
			return s3Mock, nil
		}

		funcDoGetHealthcheckOK := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetFailingHTTPSerer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return failingServerMock
		}

		Convey("Given that initialising vault returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetVaultFunc:      funcDoGetVaultErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errVault)
				So(svcList.KafkaConsumerPublished, ShouldBeFalse)
				So(svcList.S3, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising kafka consumer returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:     funcDoGetHTTPServerNil,
				DoGetVaultFunc:          funcDoGetVaultOK,
				DoGetImageAPIClientFunc: funcDoGetImageAPIClientFuncOK,
				DoGetKafkaConsumerFunc:  funcDoGetKafkaConsumerErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errKafkaConsumer)
				So(svcList.KafkaConsumerPublished, ShouldBeFalse)
				So(svcList.S3, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising the S3 clients returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:     funcDoGetHTTPServerNil,
				DoGetVaultFunc:          funcDoGetVaultOK,
				DoGetImageAPIClientFunc: funcDoGetImageAPIClientFuncOK,
				DoGetKafkaConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetS3ClientFunc:       funcDoGetS3ClientFuncErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errS3)
				So(svcList.KafkaConsumerPublished, ShouldBeTrue)
				So(svcList.S3, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:     funcDoGetHTTPServerNil,
				DoGetVaultFunc:          funcDoGetVaultOK,
				DoGetImageAPIClientFunc: funcDoGetImageAPIClientFuncOK,
				DoGetKafkaConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetS3ClientFunc:       funcDoGetS3ClientFuncOK,
				DoGetHealthCheckFunc:    funcDoGetHealthcheckErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.KafkaConsumerPublished, ShouldBeTrue)
				So(svcList.S3, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {

			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &serviceMock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:     funcDoGetHTTPServerNil,
				DoGetVaultFunc:          funcDoGetVaultOK,
				DoGetImageAPIClientFunc: funcDoGetImageAPIClientFuncOK,
				DoGetKafkaConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetS3ClientFunc:       funcDoGetS3ClientFuncOK,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddheckFail.Error()))
				So(svcList.HealthCheck, ShouldBeTrue)
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 3)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Vault")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Image API")
				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "Kafka Consumer")
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:     funcDoGetHTTPServer,
				DoGetVaultFunc:          funcDoGetVaultOK,
				DoGetImageAPIClientFunc: funcDoGetImageAPIClientFuncOK,
				DoGetKafkaConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetS3ClientFunc:       funcDoGetS3ClientFuncOK,
				DoGetHealthCheckFunc:    funcDoGetHealthcheckOK,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.KafkaConsumerPublished, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
			})

			Convey("The http server and healchecker start", func() {
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, ":24900")
				So(len(initMock.DoGetVaultCalls()), ShouldEqual, 1)
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that all dependencies are successfully initialised but the http server fails", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:     funcDoGetFailingHTTPSerer,
				DoGetVaultFunc:          funcDoGetVaultOK,
				DoGetImageAPIClientFunc: funcDoGetImageAPIClientFuncOK,
				DoGetKafkaConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetS3ClientFunc:       funcDoGetS3ClientFuncOK,
				DoGetHealthCheckFunc:    funcDoGetHealthcheckOK,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			Convey("Then the error is returned in the error channel", func() {
				sErr := <-svcErrors
				So(sErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
				So(len(failingServerMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})
	})

}

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		serverStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mock.HTTPServerMock{
			// ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		// consumer group Close will fail if healthcheck or http server are not stopped
		kafkaConsumerMock := &kafkatest.IConsumerGroupMock{
			CloseFunc: func(ctx context.Context) error {
				if !hcStopped || !serverStopped {
					return errors.New("Kafka Consumer stopped before healthcheck or HTTP server")
				}
				return nil
			},
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
					return serverMock
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
					return kafkaConsumerMock, nil
				},
			}

			svcList := service.NewServiceList(initMock)
			svcList.HealthCheck = true
			svcList.KafkaConsumerPublished = true
			svc := service.Service{
				Config:        cfg,
				ServiceList:   svcList,
				Server:        serverMock,
				HealthCheck:   hcMock,
				KafkaConsumer: kafkaConsumerMock,
			}

			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(kafkaConsumerMock.CloseCalls()), ShouldEqual, 1)

		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {

			failingserverMock := &mock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			failingKafkaConsumerMock := &kafkatest.IConsumerGroupMock{
				CloseFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop Kafka Consumer")
				},
			}

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
					return failingserverMock
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
					return failingKafkaConsumerMock, nil
				},
			}

			svcList := service.NewServiceList(initMock)
			svcList.HealthCheck = true
			svcList.KafkaConsumerPublished = true
			svc := service.Service{
				Config:        cfg,
				ServiceList:   svcList,
				Server:        failingserverMock,
				HealthCheck:   hcMock,
				KafkaConsumer: failingKafkaConsumerMock,
			}

			err := svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(failingKafkaConsumerMock.CloseCalls()), ShouldEqual, 1)
		})
	})
}
