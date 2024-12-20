package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/dp-static-file-publisher/file"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	eventMock "github.com/ONSdigital/dp-static-file-publisher/event/mock"
	fileMock "github.com/ONSdigital/dp-static-file-publisher/file/mock"
	"github.com/ONSdigital/dp-static-file-publisher/service"
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
	errKafkaConsumer = errors.New("Kafka consumer error")
	errS3            = errors.New("S3 session error")
	errServer        = errors.New("HTTP Server error")
	errHealthcheck   = errors.New("healthCheck error")
)

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetKafkaConsumerErr = func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
	return nil, errKafkaConsumer
}

var funcDoGetS3ClientV2FuncErr = func(awsRegion string, bucketName string) (file.S3ClientV2, error) {
	return nil, errS3
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
	return nil
}

func TestRun(t *testing.T) {
	Convey("Having a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		imageAPIClientMock := &eventMock.ImageAPIClientMock{}
		filesSvcClientMock := &fileMock.FilesServiceMock{}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		s3Session := &session.Session{}
		s3ClientV2Mock := &fileMock.S3ClientV2Mock{
			SessionFunc: func() *session.Session { return s3Session },
		}
		s3PrivateMock := &fileMock.S3ClientV2Mock{}

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

		funcDoGetImageAPIClientFuncOK := func(cfg *config.Config) event.ImageAPIClient {
			return imageAPIClientMock
		}

		funcDoGetFilesClientFuncOK := func(ctx context.Context, cfg *config.Config) file.FilesService {
			return filesSvcClientMock
		}

		funcDoGetS3ClientV2OK := func(awsRegion string, bucketName string) (file.S3ClientV2, error) {
			return s3ClientV2Mock, nil
		}

		funcDoGetS3ClientV2WithSessionOK := func(bucketName string, s *session.Session) (file.S3ClientV2, error) {
			return s3PrivateMock, nil
		}

		funcDoGetHealthcheckOK := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetFailingHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return failingServerMock
		}

		funcDoGetKafkaConsumerOK := func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
			return &serviceMock.KafkaConsumerMock{
				RegisterBatchHandlerFunc: func(ctx context.Context, h kafka.BatchHandler) error { return nil },
				StartFunc:                func() error { return nil },
			}, nil
		}

		Convey("Given that initialising the S3 client V2 returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:     funcDoGetHTTPServerNil,
				DoGetImageAPIClientFunc: funcDoGetImageAPIClientFuncOK,
				DoGetS3ClientV2Func:     funcDoGetS3ClientV2FuncErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errS3)
				So(svcList.KafkaImagePublishedConsumer, ShouldBeFalse)
				So(svcList.S3Public, ShouldBeFalse)
				So(svcList.S3Private, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising kafka consumer returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:                  funcDoGetHTTPServerNil,
				DoGetImageAPIClientFunc:              funcDoGetImageAPIClientFuncOK,
				DoGetS3ClientV2Func:                  funcDoGetS3ClientV2OK,
				DoGetS3ClientV2WithSessionFunc:       funcDoGetS3ClientV2WithSessionOK,
				DoGetKafkaImagePublishedConsumerFunc: funcDoGetKafkaConsumerErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errKafkaConsumer)
				So(svcList.KafkaImagePublishedConsumer, ShouldBeFalse)
				So(svcList.S3Public, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:                  funcDoGetHTTPServerNil,
				DoGetImageAPIClientFunc:              funcDoGetImageAPIClientFuncOK,
				DoGetKafkaImagePublishedConsumerFunc: funcDoGetKafkaConsumerOK,
				DoGetS3ClientV2Func:                  funcDoGetS3ClientV2OK,
				DoGetS3ClientV2WithSessionFunc:       funcDoGetS3ClientV2WithSessionOK,
				DoGetHealthCheckFunc:                 funcDoGetHealthcheckErr,
				DoGetKafkaFilePublishedConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetFilesServiceFunc:                funcDoGetFilesClientFuncOK,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.KafkaImagePublishedConsumer, ShouldBeTrue)
				So(svcList.S3Public, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising FilePublishedConsumer returns an error", func() {
			expectedError := errors.New("FilePublishedConsumer failed to initialise")

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:                  funcDoGetHTTPServerNil,
				DoGetImageAPIClientFunc:              funcDoGetImageAPIClientFuncOK,
				DoGetKafkaImagePublishedConsumerFunc: funcDoGetKafkaConsumerOK,
				DoGetS3ClientV2Func:                  funcDoGetS3ClientV2OK,
				DoGetS3ClientV2WithSessionFunc:       funcDoGetS3ClientV2WithSessionOK,
				DoGetHealthCheckFunc:                 funcDoGetHealthcheckOK,
				DoGetKafkaFilePublishedConsumerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
					return nil, expectedError
				},
				DoGetFilesServiceFunc: funcDoGetFilesClientFuncOK,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, expectedError)
				So(svcList.KafkaImagePublishedConsumer, ShouldBeTrue)
				So(svcList.S3Public, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.KafkaFilePublishedConsumer, ShouldBeFalse)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {

			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &serviceMock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:                  funcDoGetHTTPServerNil,
				DoGetImageAPIClientFunc:              funcDoGetImageAPIClientFuncOK,
				DoGetKafkaImagePublishedConsumerFunc: funcDoGetKafkaConsumerOK,
				DoGetS3ClientV2Func:                  funcDoGetS3ClientV2OK,
				DoGetS3ClientV2WithSessionFunc:       funcDoGetS3ClientV2WithSessionOK,
				DoGetKafkaFilePublishedConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetFilesServiceFunc:                funcDoGetFilesClientFuncOK,
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
				So(hcMockAddFail.AddCheckCalls(), ShouldHaveLength, 6)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Image API")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Kafka ImagePublished Consumer")
				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "Kafka FilePublished Consumer")
				So(hcMockAddFail.AddCheckCalls()[3].Name, ShouldResemble, "S3 Public")
				So(hcMockAddFail.AddCheckCalls()[4].Name, ShouldResemble, "S3 Private")
				So(hcMockAddFail.AddCheckCalls()[5].Name, ShouldResemble, "Files Service")
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:                  funcDoGetHTTPServer,
				DoGetImageAPIClientFunc:              funcDoGetImageAPIClientFuncOK,
				DoGetKafkaImagePublishedConsumerFunc: funcDoGetKafkaConsumerOK,
				DoGetS3ClientV2Func:                  funcDoGetS3ClientV2OK,
				DoGetS3ClientV2WithSessionFunc:       funcDoGetS3ClientV2WithSessionOK,
				DoGetHealthCheckFunc:                 funcDoGetHealthcheckOK,
				DoGetKafkaFilePublishedConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetFilesServiceFunc:                funcDoGetFilesClientFuncOK,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.KafkaImagePublishedConsumer, ShouldBeTrue)
				So(svcList.S3Public, ShouldBeTrue)
				So(svcList.S3Private, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.FilesService, ShouldBeTrue)

				Convey("And all healthcheck checks are registered", func() {
					So(hcMock.AddCheckCalls(), ShouldHaveLength, 6)
					So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Image API")
					So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka ImagePublished Consumer")
					So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Kafka FilePublished Consumer")
					So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "S3 Public")
					So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "S3 Private")
					So(hcMock.AddCheckCalls()[5].Name, ShouldResemble, "Files Service")
				})
			})

			Convey("The http server and healchecker start", func() {
				So(initMock.DoGetHTTPServerCalls(), ShouldHaveLength, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, ":24900")
				So(hcMock.StartCalls(), ShouldHaveLength, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(serverMock.ListenAndServeCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("Given that all dependencies are successfully initialised but the http server fails", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:                  funcDoGetFailingHTTPServer,
				DoGetImageAPIClientFunc:              funcDoGetImageAPIClientFuncOK,
				DoGetKafkaImagePublishedConsumerFunc: funcDoGetKafkaConsumerOK,
				DoGetS3ClientV2Func:                  funcDoGetS3ClientV2OK,
				DoGetS3ClientV2WithSessionFunc:       funcDoGetS3ClientV2WithSessionOK,
				DoGetHealthCheckFunc:                 funcDoGetHealthcheckOK,
				DoGetKafkaFilePublishedConsumerFunc:  funcDoGetKafkaConsumerOK,
				DoGetFilesServiceFunc:                funcDoGetFilesClientFuncOK,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			Convey("Then the error is returned in the error channel", func() {
				sErr := <-svcErrors
				So(sErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
				So(failingServerMock.ListenAndServeCalls(), ShouldHaveLength, 1)
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
		serverMock := &serviceMock.HTTPServerMock{
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		// consumer group Close and StopListeningToConsumer will fail if healthcheck or http server are not stopped
		kafkaConsumerMock := &kafkatest.IConsumerGroupMock{
			StopFunc: func() error {
				if !hcStopped || !serverStopped {
					return errors.New("Kafka Consumer StopListening before healthcheck or HTTP server")
				}
				return nil
			},
			StateWaitFunc: func(state kafka.State) {},
			CloseFunc: func(ctx context.Context) error {
				if !hcStopped || !serverStopped {
					return errors.New("Kafka Consumer stopped before healthcheck or HTTP server")
				}
				return nil
			},
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
					return serverMock
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetKafkaImagePublishedConsumerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
					return kafkaConsumerMock, nil
				},
			}

			svcList := service.NewServiceList(initMock)
			svcList.HealthCheck = true
			svcList.KafkaImagePublishedConsumer = true
			svc := service.Service{
				Config:                      cfg,
				ServiceList:                 svcList,
				Server:                      serverMock,
				HealthCheck:                 hcMock,
				KafkaImagePublishedConsumer: kafkaConsumerMock,
			}

			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(kafkaConsumerMock.StopCalls(), ShouldHaveLength, 1)
			So(kafkaConsumerMock.StateWaitCalls(), ShouldHaveLength, 1)
			So(kafkaConsumerMock.CloseCalls(), ShouldHaveLength, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {

			failingserverMock := &serviceMock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			failingKafkaConsumerMock := &kafkatest.IConsumerGroupMock{
				CloseFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop Kafka Consumer")
				},
				StopFunc: func() error {
					return errors.New("Failed to stop listening to consumer")
				},
				StateWaitFunc: func(state kafka.State) {},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
					return failingserverMock
				},
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetKafkaImagePublishedConsumerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
					return failingKafkaConsumerMock, nil
				},
			}

			svcList := service.NewServiceList(initMock)
			svcList.HealthCheck = true
			svcList.KafkaImagePublishedConsumer = true
			svc := service.Service{
				Config:                      cfg,
				ServiceList:                 svcList,
				Server:                      failingserverMock,
				HealthCheck:                 hcMock,
				KafkaImagePublishedConsumer: failingKafkaConsumerMock,
			}

			err := svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(failingserverMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(failingKafkaConsumerMock.StopCalls(), ShouldHaveLength, 1)
			So(failingKafkaConsumerMock.StateWaitCalls(), ShouldHaveLength, 1)
			So(failingKafkaConsumerMock.CloseCalls(), ShouldHaveLength, 1)
		})
	})
}
