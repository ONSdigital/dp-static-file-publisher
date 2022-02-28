package steps

import (
	"context"
	"net/http"
	"time"

	dpkafka "github.com/ONSdigital/dp-kafka/v2"

	"github.com/ONSdigital/dp-kafka/v2/kafkatest"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"github.com/aws/aws-sdk-go/aws/session"
)

type fakeServiceContainer struct {
	server *dphttp.Server
}

func (e *fakeServiceContainer) DoGetHTTPServer(bindAddr string, r http.Handler) service.HTTPServer {
	e.server.Server.Addr = ":26900"
	e.server.Server.Handler = r

	return e.server
}

func (e *fakeServiceContainer) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	h := healthcheck.New(healthcheck.VersionInfo{}, time.Second, time.Second)
	return &h, nil
}

func (e *fakeServiceContainer) DoGetVault(cfg *config.Config) (event.VaultClient, error) {
	return &mock.VaultClientMock{
		CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update("OK", "Vault all good", 0)
			return nil
		},
	}, nil
}

func (e *fakeServiceContainer) DoGetImageAPIClient(cfg *config.Config) event.ImageAPIClient {
	return &mock.ImageAPIClientMock{
		CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update("OK", "Image API all good", 0)
			return nil
		},
	}
}

func (e *fakeServiceContainer) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
	return &kafkatest.IConsumerGroupMock{
		CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update("OK", "Kafka Consumer all good", 0)
			return nil
		},
		ChannelsFunc: func() *dpkafka.ConsumerGroupChannels {
			return dpkafka.CreateConsumerGroupChannels(1)
		},
	}, nil
}

func (e *fakeServiceContainer) DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
	return &mock.S3WriterMock{
		CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update("OK", "s3 Client Without Session all good", 0)
			return nil
		},
		SessionFunc: func() *session.Session {
			s, _ := session.NewSession()
			return s
		},
	}, nil
}

func (e *fakeServiceContainer) DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
	return &mock.S3ReaderMock{
		CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update("OK", "s3 Client With Session all good", 0)
			return nil
		},
	}
}

func (e *fakeServiceContainer) Shutdown(ctx context.Context) error {
	return nil
}
