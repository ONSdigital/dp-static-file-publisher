package steps

import (
	"context"
	"fmt"
	"net/http"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	dps3v2 "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	fmock "github.com/ONSdigital/dp-static-file-publisher/file/mock"

	s3client "github.com/ONSdigital/dp-s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	vault "github.com/ONSdigital/dp-vault"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"github.com/aws/aws-sdk-go/aws/session"
)

type fakeServiceContainer struct {
	server     *dphttp.Server
	decryptReq map[string]string
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
	v, err := vault.CreateClient(cfg.VaultToken, cfg.VaultAddress, cfg.VaultRetries)
	if err != nil {
		fmt.Println(err.Error())
	}

	return v, nil
}

func (e *fakeServiceContainer) DoGetImageAPIClient(cfg *config.Config) event.ImageAPIClient {
	return &mock.ImageAPIClientMock{
		CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update("OK", "Image API all good", 0)
			return nil
		},
	}
}

func (e *fakeServiceContainer) DoGetFilesService(ctx context.Context, cfg *config.Config) file.FilesService {
	return &fmock.FilesServiceMock{
		CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update("OK", "Files Service API all good", 0)
			return nil
		},
		MarkFileDecryptedFunc: func(ctx context.Context, path string, etag string) error {
			e.decryptReq[path] = "DECRYPTED"
			return nil
		},
	}
}

func (e *fakeServiceContainer) DoGetKafkaImagePublishedConsumer(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
	kafkaOffset := kafka.OffsetOldest

	gc := kafka.ConsumerGroupConfig{
		KafkaVersion:      &cfg.KafkaVersion,
		Offset:            &kafkaOffset,
		MinBrokersHealthy: &cfg.KafkaMinimumHealthyBrokers,
		Topic:             cfg.ImageFilePublishedTopic,
		GroupName:         cfg.ConsumerGroup,
		BrokerAddrs:       cfg.KafkaAddr,
	}

	return kafka.NewConsumerGroup(ctx, &gc)
}

func (e *fakeServiceContainer) DoGetKafkaFilePublishedConsumer(ctx context.Context, cfg *config.Config) (service.KafkaConsumer, error) {
	kafkaOffset := kafka.OffsetOldest

	gc := kafka.ConsumerGroupConfig{
		KafkaVersion:      &cfg.KafkaVersion,
		Offset:            &kafkaOffset,
		MinBrokersHealthy: &cfg.KafkaMinimumHealthyBrokers,
		Topic:             cfg.StaticFilePublishedTopic,
		GroupName:         cfg.ConsumerGroup,
		BrokerAddrs:       cfg.KafkaAddr,
	}

	return kafka.NewConsumerGroup(ctx, &gc)
}

func (e *fakeServiceContainer) DoGetS3Client(awsRegion, bucketName string, encryptionEnabled bool) (event.S3Writer, error) {
	s, _ := session.NewSession(&aws.Config{
		Endpoint:         aws.String(localStackHost),
		Region:           aws.String(awsRegion),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
	})

	return s3client.NewUploaderWithSession(bucketName, encryptionEnabled, s), nil
}

func (e *fakeServiceContainer) DoGetS3ClientV2(awsRegion, bucketName string) (file.S3ClientV2, error) {
	s, err := session.NewSession(&aws.Config{
		Region:           aws.String(awsRegion),
		Endpoint:         aws.String(localStackHost),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
	})

	if err != nil {
		return nil, err
	}

	return dps3v2.NewClientWithSession(bucketName, s), nil
}

func (e *fakeServiceContainer) DoGetS3ClientWithSession(bucketName string, encryptionEnabled bool, s *session.Session) event.S3Reader {
	return s3client.NewClientWithSession(bucketName, encryptionEnabled, s)
}

func (e *fakeServiceContainer) Shutdown(ctx context.Context) error {
	return nil
}
