// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"github.com/aws/aws-sdk-go/aws/session"
	"net/http"
	"sync"
)

var (
	lockInitialiserMockDoGetHTTPServer          sync.RWMutex
	lockInitialiserMockDoGetHealthCheck         sync.RWMutex
	lockInitialiserMockDoGetImageAPIClient      sync.RWMutex
	lockInitialiserMockDoGetKafkaConsumer       sync.RWMutex
	lockInitialiserMockDoGetS3Client            sync.RWMutex
	lockInitialiserMockDoGetS3ClientWithSession sync.RWMutex
	lockInitialiserMockDoGetVault               sync.RWMutex
)

// Ensure, that InitialiserMock does implement service.Initialiser.
// If this is not the case, regenerate this file with moq.
var _ service.Initialiser = &InitialiserMock{}

// InitialiserMock is a mock implementation of service.Initialiser.
//
//     func TestSomethingThatUsesInitialiser(t *testing.T) {
//
//         // make and configure a mocked service.Initialiser
//         mockedInitialiser := &InitialiserMock{
//             DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
// 	               panic("mock out the DoGetHTTPServer method")
//             },
//             DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 	               panic("mock out the DoGetHealthCheck method")
//             },
//             DoGetImageAPIClientFunc: func(imageAPIURL string) service.ImageAPIClient {
// 	               panic("mock out the DoGetImageAPIClient method")
//             },
//             DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
// 	               panic("mock out the DoGetKafkaConsumer method")
//             },
//             DoGetS3ClientFunc: func(awsRegion string, bucketName string, encryptionEnabled bool) (event.S3Client, error) {
// 	               panic("mock out the DoGetS3Client method")
//             },
//             DoGetS3ClientWithSessionFunc: func(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Client {
// 	               panic("mock out the DoGetS3ClientWithSession method")
//             },
//             DoGetVaultFunc: func(vaultToken string, vaultAddress string, retries int) (service.VaultClient, error) {
// 	               panic("mock out the DoGetVault method")
//             },
//         }
//
//         // use mockedInitialiser in code that requires service.Initialiser
//         // and then make assertions.
//
//     }
type InitialiserMock struct {
	// DoGetHTTPServerFunc mocks the DoGetHTTPServer method.
	DoGetHTTPServerFunc func(bindAddr string, router http.Handler) service.HTTPServer

	// DoGetHealthCheckFunc mocks the DoGetHealthCheck method.
	DoGetHealthCheckFunc func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error)

	// DoGetImageAPIClientFunc mocks the DoGetImageAPIClient method.
	DoGetImageAPIClientFunc func(imageAPIURL string) service.ImageAPIClient

	// DoGetKafkaConsumerFunc mocks the DoGetKafkaConsumer method.
	DoGetKafkaConsumerFunc func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error)

	// DoGetS3ClientFunc mocks the DoGetS3Client method.
	DoGetS3ClientFunc func(awsRegion string, bucketName string, encryptionEnabled bool) (event.S3Client, error)

	// DoGetS3ClientWithSessionFunc mocks the DoGetS3ClientWithSession method.
	DoGetS3ClientWithSessionFunc func(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Client

	// DoGetVaultFunc mocks the DoGetVault method.
	DoGetVaultFunc func(vaultToken string, vaultAddress string, retries int) (service.VaultClient, error)

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
			// ImageAPIURL is the imageAPIURL argument value.
			ImageAPIURL string
		}
		// DoGetKafkaConsumer holds details about calls to the DoGetKafkaConsumer method.
		DoGetKafkaConsumer []struct {
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
			// VaultToken is the vaultToken argument value.
			VaultToken string
			// VaultAddress is the vaultAddress argument value.
			VaultAddress string
			// Retries is the retries argument value.
			Retries int
		}
	}
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
	lockInitialiserMockDoGetHTTPServer.Lock()
	mock.calls.DoGetHTTPServer = append(mock.calls.DoGetHTTPServer, callInfo)
	lockInitialiserMockDoGetHTTPServer.Unlock()
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
	lockInitialiserMockDoGetHTTPServer.RLock()
	calls = mock.calls.DoGetHTTPServer
	lockInitialiserMockDoGetHTTPServer.RUnlock()
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
	lockInitialiserMockDoGetHealthCheck.Lock()
	mock.calls.DoGetHealthCheck = append(mock.calls.DoGetHealthCheck, callInfo)
	lockInitialiserMockDoGetHealthCheck.Unlock()
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
	lockInitialiserMockDoGetHealthCheck.RLock()
	calls = mock.calls.DoGetHealthCheck
	lockInitialiserMockDoGetHealthCheck.RUnlock()
	return calls
}

// DoGetImageAPIClient calls DoGetImageAPIClientFunc.
func (mock *InitialiserMock) DoGetImageAPIClient(imageAPIURL string) service.ImageAPIClient {
	if mock.DoGetImageAPIClientFunc == nil {
		panic("InitialiserMock.DoGetImageAPIClientFunc: method is nil but Initialiser.DoGetImageAPIClient was just called")
	}
	callInfo := struct {
		ImageAPIURL string
	}{
		ImageAPIURL: imageAPIURL,
	}
	lockInitialiserMockDoGetImageAPIClient.Lock()
	mock.calls.DoGetImageAPIClient = append(mock.calls.DoGetImageAPIClient, callInfo)
	lockInitialiserMockDoGetImageAPIClient.Unlock()
	return mock.DoGetImageAPIClientFunc(imageAPIURL)
}

// DoGetImageAPIClientCalls gets all the calls that were made to DoGetImageAPIClient.
// Check the length with:
//     len(mockedInitialiser.DoGetImageAPIClientCalls())
func (mock *InitialiserMock) DoGetImageAPIClientCalls() []struct {
	ImageAPIURL string
} {
	var calls []struct {
		ImageAPIURL string
	}
	lockInitialiserMockDoGetImageAPIClient.RLock()
	calls = mock.calls.DoGetImageAPIClient
	lockInitialiserMockDoGetImageAPIClient.RUnlock()
	return calls
}

// DoGetKafkaConsumer calls DoGetKafkaConsumerFunc.
func (mock *InitialiserMock) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
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
	lockInitialiserMockDoGetKafkaConsumer.Lock()
	mock.calls.DoGetKafkaConsumer = append(mock.calls.DoGetKafkaConsumer, callInfo)
	lockInitialiserMockDoGetKafkaConsumer.Unlock()
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
	lockInitialiserMockDoGetKafkaConsumer.RLock()
	calls = mock.calls.DoGetKafkaConsumer
	lockInitialiserMockDoGetKafkaConsumer.RUnlock()
	return calls
}

// DoGetS3Client calls DoGetS3ClientFunc.
func (mock *InitialiserMock) DoGetS3Client(awsRegion string, bucketName string, encryptionEnabled bool) (event.S3Client, error) {
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
	lockInitialiserMockDoGetS3Client.Lock()
	mock.calls.DoGetS3Client = append(mock.calls.DoGetS3Client, callInfo)
	lockInitialiserMockDoGetS3Client.Unlock()
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
	lockInitialiserMockDoGetS3Client.RLock()
	calls = mock.calls.DoGetS3Client
	lockInitialiserMockDoGetS3Client.RUnlock()
	return calls
}

// DoGetS3ClientWithSession calls DoGetS3ClientWithSessionFunc.
func (mock *InitialiserMock) DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Client {
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
	lockInitialiserMockDoGetS3ClientWithSession.Lock()
	mock.calls.DoGetS3ClientWithSession = append(mock.calls.DoGetS3ClientWithSession, callInfo)
	lockInitialiserMockDoGetS3ClientWithSession.Unlock()
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
	lockInitialiserMockDoGetS3ClientWithSession.RLock()
	calls = mock.calls.DoGetS3ClientWithSession
	lockInitialiserMockDoGetS3ClientWithSession.RUnlock()
	return calls
}

// DoGetVault calls DoGetVaultFunc.
func (mock *InitialiserMock) DoGetVault(vaultToken string, vaultAddress string, retries int) (service.VaultClient, error) {
	if mock.DoGetVaultFunc == nil {
		panic("InitialiserMock.DoGetVaultFunc: method is nil but Initialiser.DoGetVault was just called")
	}
	callInfo := struct {
		VaultToken   string
		VaultAddress string
		Retries      int
	}{
		VaultToken:   vaultToken,
		VaultAddress: vaultAddress,
		Retries:      retries,
	}
	lockInitialiserMockDoGetVault.Lock()
	mock.calls.DoGetVault = append(mock.calls.DoGetVault, callInfo)
	lockInitialiserMockDoGetVault.Unlock()
	return mock.DoGetVaultFunc(vaultToken, vaultAddress, retries)
}

// DoGetVaultCalls gets all the calls that were made to DoGetVault.
// Check the length with:
//     len(mockedInitialiser.DoGetVaultCalls())
func (mock *InitialiserMock) DoGetVaultCalls() []struct {
	VaultToken   string
	VaultAddress string
	Retries      int
} {
	var calls []struct {
		VaultToken   string
		VaultAddress string
		Retries      int
	}
	lockInitialiserMockDoGetVault.RLock()
	calls = mock.calls.DoGetVault
	lockInitialiserMockDoGetVault.RUnlock()
	return calls
}
