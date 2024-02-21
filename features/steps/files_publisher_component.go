package steps

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/ONSdigital/dp-static-file-publisher/config"

	componenttest "github.com/ONSdigital/dp-component-test"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-static-file-publisher/service"
	"github.com/ONSdigital/log.go/v2/log"
)

type FilePublisherComponent struct {
	DpHttpServer *dphttp.Server
	svc          *service.Service
	svcList      service.Initialiser
	ApiFeature   *componenttest.APIFeature
	errChan      chan error
	config       *config.Config
	session      *session.Session
	request      map[string]string
}

const (
	localStackHost = "http://localstack:4566"
)

func NewFilePublisherComponent() *FilePublisherComponent {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	d := &FilePublisherComponent{
		DpHttpServer: s,
		errChan:      make(chan error),
	}

	log.Namespace = "dp-static-file-publisher"
	d.request = make(map[string]string, 10)
	d.svcList = &fakeServiceContainer{s, d.request}
	d.config, _ = config.Get()
	d.session, _ = session.NewSession(&aws.Config{
		Endpoint:         aws.String(localStackHost),
		Region:           aws.String(d.config.AwsRegion),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", "")})

	return d
}

func (c *FilePublisherComponent) Initialiser() (http.Handler, error) {
	cfg, _ := config.Get()
	c.svc, _ = service.Run(context.Background(), cfg, service.NewServiceList(c.svcList), "0", "0", "1.0.0", c.errChan)
	time.Sleep(30 * time.Second)
	return c.DpHttpServer.Handler, nil
}

func (c *FilePublisherComponent) Reset() {
	cfg, _ := config.Get()
	s, _ := session.NewSession(&aws.Config{
		Endpoint:         aws.String(localStackHost),
		Region:           aws.String(cfg.AwsRegion),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
	})

	s3client := s3.New(s)

	err := s3manager.NewBatchDeleteWithClient(s3client).Delete(
		aws.BackgroundContext(), s3manager.NewDeleteListIterator(s3client, &s3.ListObjectsInput{
			Bucket: aws.String(cfg.PrivateBucketName),
		}))

	if err != nil {
		panic(fmt.Sprintf("Failed to empty private localstack s3: %s", err.Error()))
	}

	err = s3manager.NewBatchDeleteWithClient(s3client).Delete(
		aws.BackgroundContext(), s3manager.NewDeleteListIterator(s3client, &s3.ListObjectsInput{
			Bucket: aws.String(cfg.PublicBucketName),
		}))

	if err != nil {
		panic(fmt.Sprintf("Failed to empty public localstack s3: %s", err.Error()))
	}
}

func (c *FilePublisherComponent) Close() error {
	if c.svc != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		return c.svc.Close(ctx)
	}
	return nil
}
