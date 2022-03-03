// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"github.com/aws/aws-sdk-go/aws/session"
	"net/http"
	"sync"
)

// Ensure, that InitialiserMock does implement service.Initialiser.
// If this is not the case, regenerate this file with moq.
var _ service.Initialiser = &InitialiserMock{}

// InitialiserMock is a mock implementation of service.Initialiser.
//
// 	func TestSomethingThatUsesInitialiser(t *testing.T) {
//
// 		// make and configure a mocked service.Initialiser
// 		mockedInitialiser := &InitialiserMock{
// 			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
// 				panic("mock out the DoGetHTTPServer method")
// 			},
// 			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 				panic("mock out the DoGetHealthCheck method")
// 			},
// 			DoGetImageAPIClientFunc: func(cfg *config.Config) event.ImageAPIClient {
// 				panic("mock out the DoGetImageAPIClient method")
// 			},
// 			DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
// 				panic("mock out the DoGetKafkaConsumer method")
// 			},
// 			DoGetKafkaV3ConsumerFunc: func(ctx context.Context, cfg *config.Config) (service.KafkaConsumerV3, error) {
// 				panic("mock out the DoGetKafkaV3Consumer method")
// 			},
// 			DoGetS3ClientFunc: func(awsRegion string, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
// 				panic("mock out the DoGetS3Client method")
// 			},
// 			DoGetS3ClientV2Func: func(awsRegion string, bucketName string) (file.S3ClientV2, error) {
// 				panic("mock out the DoGetS3ClientV2 method")
// 			},
// 			DoGetS3ClientWithSessionFunc: func(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
// 				panic("mock out the DoGetS3ClientWithSession method")
// 			},
// 			DoGetVaultFunc: func(cfg *config.Config) (event.VaultClient, error) {
// 				panic("mock out the DoGetVault method")
// 			},
// 		}
//
// 		// use mockedInitialiser in code that requires service.Initialiser
// 		// and then make assertions.
//
// 	}
type InitialiserMock struct {
	// DoGetHTTPServerFunc mocks the DoGetHTTPServer method.
	DoGetHTTPServerFunc func(bindAddr string, router http.Handler) service.HTTPServer

	// DoGetHealthCheckFunc mocks the DoGetHealthCheck method.
	DoGetHealthCheckFunc func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error)

	// DoGetImageAPIClientFunc mocks the DoGetImageAPIClient method.
	DoGetImageAPIClientFunc func(cfg *config.Config) event.ImageAPIClient

	// DoGetKafkaConsumerFunc mocks the DoGetKafkaConsumer method.
	DoGetKafkaConsumerFunc func(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error)

	// DoGetKafkaV3ConsumerFunc mocks the DoGetKafkaV3Consumer method.
	DoGetKafkaV3ConsumerFunc func(ctx context.Context, cfg *config.Config) (service.KafkaConsumerV3, error)

	// DoGetS3ClientFunc mocks the DoGetS3Client method.
	DoGetS3ClientFunc func(awsRegion string, bucketName string, encryptionEnabled bool) (event.S3Writer, error)

	// DoGetS3ClientV2Func mocks the DoGetS3ClientV2 method.
	DoGetS3ClientV2Func func(awsRegion string, bucketName string) (file.S3ClientV2, error)

	// DoGetS3ClientWithSessionFunc mocks the DoGetS3ClientWithSession method.
	DoGetS3ClientWithSessionFunc func(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader

	// DoGetVaultFunc mocks the DoGetVault method.
	DoGetVaultFunc func(cfg *config.Config) (event.VaultClient, error)

	// calls tracks calls to the methods.
	calls struct {
		// DoGetHTTPServer holds details about calls to the DoGetHTTPServer method.
		DoGetHTTPServer []struct {
			// BindAddr is the bindAddr argument value.
			BindAddr string
			// Router is the router argument value.
			Router http.Handler
		}
		// DoGetHealthCheck holds details about calls to the DoGetHealthCheck method.
		DoGetHealthCheck []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
			// BuildTime is the buildTime argument value.
			BuildTime string
			// GitCommit is the gitCommit argument value.
			GitCommit string
			// Version is the version argument value.
			Version string
		}
		// DoGetImageAPIClient holds details about calls to the DoGetImageAPIClient method.
		DoGetImageAPIClient []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetKafkaConsumer holds details about calls to the DoGetKafkaConsumer method.
		DoGetKafkaConsumer []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetKafkaV3Consumer holds details about calls to the DoGetKafkaV3Consumer method.
		DoGetKafkaV3Consumer []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetS3Client holds details about calls to the DoGetS3Client method.
		DoGetS3Client []struct {
			// AwsRegion is the awsRegion argument value.
			AwsRegion string
			// BucketName is the bucketName argument value.
			BucketName string
			// EncryptionEnabled is the encryptionEnabled argument value.
			EncryptionEnabled bool
		}
		// DoGetS3ClientV2 holds details about calls to the DoGetS3ClientV2 method.
		DoGetS3ClientV2 []struct {
			// AwsRegion is the awsRegion argument value.
			AwsRegion string
			// BucketName is the bucketName argument value.
			BucketName string
		}
		// DoGetS3ClientWithSession holds details about calls to the DoGetS3ClientWithSession method.
		DoGetS3ClientWithSession []struct {
			// BucketName is the bucketName argument value.
			BucketName string
			// EncryptionEnabled is the encryptionEnabled argument value.
			EncryptionEnabled bool
			// S is the s argument value.
			S *session.Session
		}
		// DoGetVault holds details about calls to the DoGetVault method.
		DoGetVault []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
	}
	lockDoGetHTTPServer          sync.RWMutex
	lockDoGetHealthCheck         sync.RWMutex
	lockDoGetImageAPIClient      sync.RWMutex
	lockDoGetKafkaConsumer       sync.RWMutex
	lockDoGetKafkaV3Consumer     sync.RWMutex
	lockDoGetS3Client            sync.RWMutex
	lockDoGetS3ClientV2          sync.RWMutex
	lockDoGetS3ClientWithSession sync.RWMutex
	lockDoGetVault               sync.RWMutex
}

// DoGetHTTPServer calls DoGetHTTPServerFunc.
func (mock *InitialiserMock) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	if mock.DoGetHTTPServerFunc == nil {
		panic("InitialiserMock.DoGetHTTPServerFunc: method is nil but Initialiser.DoGetHTTPServer was just called")
	}
	callInfo := struct {
		BindAddr string
		Router   http.Handler
	}{
		BindAddr: bindAddr,
		Router:   router,
	}
	mock.lockDoGetHTTPServer.Lock()
	mock.calls.DoGetHTTPServer = append(mock.calls.DoGetHTTPServer, callInfo)
	mock.lockDoGetHTTPServer.Unlock()
	return mock.DoGetHTTPServerFunc(bindAddr, router)
}

// DoGetHTTPServerCalls gets all the calls that were made to DoGetHTTPServer.
// Check the length with:
//     len(mockedInitialiser.DoGetHTTPServerCalls())
func (mock *InitialiserMock) DoGetHTTPServerCalls() []struct {
	BindAddr string
	Router   http.Handler
} {
	var calls []struct {
		BindAddr string
		Router   http.Handler
	}
	mock.lockDoGetHTTPServer.RLock()
	calls = mock.calls.DoGetHTTPServer
	mock.lockDoGetHTTPServer.RUnlock()
	return calls
}

// DoGetHealthCheck calls DoGetHealthCheckFunc.
func (mock *InitialiserMock) DoGetHealthCheck(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	if mock.DoGetHealthCheckFunc == nil {
		panic("InitialiserMock.DoGetHealthCheckFunc: method is nil but Initialiser.DoGetHealthCheck was just called")
	}
	callInfo := struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}{
		Cfg:       cfg,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		Version:   version,
	}
	mock.lockDoGetHealthCheck.Lock()
	mock.calls.DoGetHealthCheck = append(mock.calls.DoGetHealthCheck, callInfo)
	mock.lockDoGetHealthCheck.Unlock()
	return mock.DoGetHealthCheckFunc(cfg, buildTime, gitCommit, version)
}

// DoGetHealthCheckCalls gets all the calls that were made to DoGetHealthCheck.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthCheckCalls())
func (mock *InitialiserMock) DoGetHealthCheckCalls() []struct {
	Cfg       *config.Config
	BuildTime string
	GitCommit string
	Version   string
} {
	var calls []struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}
	mock.lockDoGetHealthCheck.RLock()
	calls = mock.calls.DoGetHealthCheck
	mock.lockDoGetHealthCheck.RUnlock()
	return calls
}

// DoGetImageAPIClient calls DoGetImageAPIClientFunc.
func (mock *InitialiserMock) DoGetImageAPIClient(cfg *config.Config) event.ImageAPIClient {
	if mock.DoGetImageAPIClientFunc == nil {
		panic("InitialiserMock.DoGetImageAPIClientFunc: method is nil but Initialiser.DoGetImageAPIClient was just called")
	}
	callInfo := struct {
		Cfg *config.Config
	}{
		Cfg: cfg,
	}
	mock.lockDoGetImageAPIClient.Lock()
	mock.calls.DoGetImageAPIClient = append(mock.calls.DoGetImageAPIClient, callInfo)
	mock.lockDoGetImageAPIClient.Unlock()
	return mock.DoGetImageAPIClientFunc(cfg)
}

// DoGetImageAPIClientCalls gets all the calls that were made to DoGetImageAPIClient.
// Check the length with:
//     len(mockedInitialiser.DoGetImageAPIClientCalls())
func (mock *InitialiserMock) DoGetImageAPIClientCalls() []struct {
	Cfg *config.Config
} {
	var calls []struct {
		Cfg *config.Config
	}
	mock.lockDoGetImageAPIClient.RLock()
	calls = mock.calls.DoGetImageAPIClient
	mock.lockDoGetImageAPIClient.RUnlock()
	return calls
}

// DoGetKafkaConsumer calls DoGetKafkaConsumerFunc.
func (mock *InitialiserMock) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
	if mock.DoGetKafkaConsumerFunc == nil {
		panic("InitialiserMock.DoGetKafkaConsumerFunc: method is nil but Initialiser.DoGetKafkaConsumer was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetKafkaConsumer.Lock()
	mock.calls.DoGetKafkaConsumer = append(mock.calls.DoGetKafkaConsumer, callInfo)
	mock.lockDoGetKafkaConsumer.Unlock()
	return mock.DoGetKafkaConsumerFunc(ctx, cfg)
}

// DoGetKafkaConsumerCalls gets all the calls that were made to DoGetKafkaConsumer.
// Check the length with:
//     len(mockedInitialiser.DoGetKafkaConsumerCalls())
func (mock *InitialiserMock) DoGetKafkaConsumerCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetKafkaConsumer.RLock()
	calls = mock.calls.DoGetKafkaConsumer
	mock.lockDoGetKafkaConsumer.RUnlock()
	return calls
}

// DoGetKafkaV3Consumer calls DoGetKafkaV3ConsumerFunc.
func (mock *InitialiserMock) DoGetKafkaV3Consumer(ctx context.Context, cfg *config.Config) (service.KafkaConsumerV3, error) {
	if mock.DoGetKafkaV3ConsumerFunc == nil {
		panic("InitialiserMock.DoGetKafkaV3ConsumerFunc: method is nil but Initialiser.DoGetKafkaV3Consumer was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetKafkaV3Consumer.Lock()
	mock.calls.DoGetKafkaV3Consumer = append(mock.calls.DoGetKafkaV3Consumer, callInfo)
	mock.lockDoGetKafkaV3Consumer.Unlock()
	return mock.DoGetKafkaV3ConsumerFunc(ctx, cfg)
}

// DoGetKafkaV3ConsumerCalls gets all the calls that were made to DoGetKafkaV3Consumer.
// Check the length with:
//     len(mockedInitialiser.DoGetKafkaV3ConsumerCalls())
func (mock *InitialiserMock) DoGetKafkaV3ConsumerCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetKafkaV3Consumer.RLock()
	calls = mock.calls.DoGetKafkaV3Consumer
	mock.lockDoGetKafkaV3Consumer.RUnlock()
	return calls
}

// DoGetS3Client calls DoGetS3ClientFunc.
func (mock *InitialiserMock) DoGetS3Client(awsRegion string, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
	if mock.DoGetS3ClientFunc == nil {
		panic("InitialiserMock.DoGetS3ClientFunc: method is nil but Initialiser.DoGetS3Client was just called")
	}
	callInfo := struct {
		AwsRegion         string
		BucketName        string
		EncryptionEnabled bool
	}{
		AwsRegion:         awsRegion,
		BucketName:        bucketName,
		EncryptionEnabled: encryptionEnabled,
	}
	mock.lockDoGetS3Client.Lock()
	mock.calls.DoGetS3Client = append(mock.calls.DoGetS3Client, callInfo)
	mock.lockDoGetS3Client.Unlock()
	return mock.DoGetS3ClientFunc(awsRegion, bucketName, encryptionEnabled)
}

// DoGetS3ClientCalls gets all the calls that were made to DoGetS3Client.
// Check the length with:
//     len(mockedInitialiser.DoGetS3ClientCalls())
func (mock *InitialiserMock) DoGetS3ClientCalls() []struct {
	AwsRegion         string
	BucketName        string
	EncryptionEnabled bool
} {
	var calls []struct {
		AwsRegion         string
		BucketName        string
		EncryptionEnabled bool
	}
	mock.lockDoGetS3Client.RLock()
	calls = mock.calls.DoGetS3Client
	mock.lockDoGetS3Client.RUnlock()
	return calls
}

// DoGetS3ClientV2 calls DoGetS3ClientV2Func.
func (mock *InitialiserMock) DoGetS3ClientV2(awsRegion string, bucketName string) (file.S3ClientV2, error) {
	if mock.DoGetS3ClientV2Func == nil {
		panic("InitialiserMock.DoGetS3ClientV2Func: method is nil but Initialiser.DoGetS3ClientV2 was just called")
	}
	callInfo := struct {
		AwsRegion  string
		BucketName string
	}{
		AwsRegion:  awsRegion,
		BucketName: bucketName,
	}
	mock.lockDoGetS3ClientV2.Lock()
	mock.calls.DoGetS3ClientV2 = append(mock.calls.DoGetS3ClientV2, callInfo)
	mock.lockDoGetS3ClientV2.Unlock()
	return mock.DoGetS3ClientV2Func(awsRegion, bucketName)
}

// DoGetS3ClientV2Calls gets all the calls that were made to DoGetS3ClientV2.
// Check the length with:
//     len(mockedInitialiser.DoGetS3ClientV2Calls())
func (mock *InitialiserMock) DoGetS3ClientV2Calls() []struct {
	AwsRegion  string
	BucketName string
} {
	var calls []struct {
		AwsRegion  string
		BucketName string
	}
	mock.lockDoGetS3ClientV2.RLock()
	calls = mock.calls.DoGetS3ClientV2
	mock.lockDoGetS3ClientV2.RUnlock()
	return calls
}

// DoGetS3ClientWithSession calls DoGetS3ClientWithSessionFunc.
func (mock *InitialiserMock) DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
	if mock.DoGetS3ClientWithSessionFunc == nil {
		panic("InitialiserMock.DoGetS3ClientWithSessionFunc: method is nil but Initialiser.DoGetS3ClientWithSession was just called")
	}
	callInfo := struct {
		BucketName        string
		EncryptionEnabled bool
		S                 *session.Session
	}{
		BucketName:        bucketName,
		EncryptionEnabled: encryptionEnabled,
		S:                 s,
	}
	mock.lockDoGetS3ClientWithSession.Lock()
	mock.calls.DoGetS3ClientWithSession = append(mock.calls.DoGetS3ClientWithSession, callInfo)
	mock.lockDoGetS3ClientWithSession.Unlock()
	return mock.DoGetS3ClientWithSessionFunc(bucketName, encryptionEnabled, s)
}

// DoGetS3ClientWithSessionCalls gets all the calls that were made to DoGetS3ClientWithSession.
// Check the length with:
//     len(mockedInitialiser.DoGetS3ClientWithSessionCalls())
func (mock *InitialiserMock) DoGetS3ClientWithSessionCalls() []struct {
	BucketName        string
	EncryptionEnabled bool
	S                 *session.Session
} {
	var calls []struct {
		BucketName        string
		EncryptionEnabled bool
		S                 *session.Session
	}
	mock.lockDoGetS3ClientWithSession.RLock()
	calls = mock.calls.DoGetS3ClientWithSession
	mock.lockDoGetS3ClientWithSession.RUnlock()
	return calls
}

// DoGetVault calls DoGetVaultFunc.
func (mock *InitialiserMock) DoGetVault(cfg *config.Config) (event.VaultClient, error) {
	if mock.DoGetVaultFunc == nil {
		panic("InitialiserMock.DoGetVaultFunc: method is nil but Initialiser.DoGetVault was just called")
	}
	callInfo := struct {
		Cfg *config.Config
	}{
		Cfg: cfg,
	}
	mock.lockDoGetVault.Lock()
	mock.calls.DoGetVault = append(mock.calls.DoGetVault, callInfo)
	mock.lockDoGetVault.Unlock()
	return mock.DoGetVaultFunc(cfg)
}

// DoGetVaultCalls gets all the calls that were made to DoGetVault.
// Check the length with:
//     len(mockedInitialiser.DoGetVaultCalls())
func (mock *InitialiserMock) DoGetVaultCalls() []struct {
	Cfg *config.Config
} {
	var calls []struct {
		Cfg *config.Config
	}
	mock.lockDoGetVault.RLock()
	calls = mock.calls.DoGetVault
	mock.lockDoGetVault.RUnlock()
	return calls
}
