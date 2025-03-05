package steps

import (
	"context"
	"net/http"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	dps3 "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	fmock "github.com/ONSdigital/dp-static-file-publisher/file/mock"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/ONSdigital/dp-static-file-publisher/service"
)

type fakeServiceContainer struct {
	server  *dphttp.Server
	moveReq map[string]string
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
		MarkFileMovedFunc: func(ctx context.Context, path string, etag string) error {
			e.moveReq[path] = "MOVED"
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

func (e *fakeServiceContainer) DoGetS3Client(ctx context.Context, awsRegion, bucketName string) (file.S3Client, error) {
	AWSConfig, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(awsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)

	if err != nil {
		return nil, err
	}

	client := dps3.NewClientWithConfig(bucketName, AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})

	return client, nil
}

func (e *fakeServiceContainer) DoGetS3ClientWithConfig(bucketName string, cfg aws.Config) (file.S3Client, error) {
	return dps3.NewClientWithConfig(bucketName, cfg), nil
}

func (e *fakeServiceContainer) Shutdown(ctx context.Context) error {
	return nil
}
