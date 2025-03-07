package steps

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-static-file-publisher/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

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
	AWSConfig    *aws.Config
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

	c, _ := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(d.config.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)

	d.AWSConfig = &c
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

	conf, _ := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)

	client := s3.NewFromConfig(conf, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})

	// delete all objects in the private bucket
	deleteObjectsInBucket(cfg.PrivateBucketName, client)

	// delete all objects in the public bucket
	deleteObjectsInBucket(cfg.PublicBucketName, client)
}

func (c *FilePublisherComponent) Close() error {
	if c.svc != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		return c.svc.Close(ctx)
	}
	return nil
}

func deleteObjectsInBucket(bucketName string, client *s3.Client) {
	listObjectInput := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}

	listObjectOutput, _ := client.ListObjects(context.Background(), listObjectInput)

	for _, object := range listObjectOutput.Contents {
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    object.Key,
		}
		_, err := client.DeleteObject(context.Background(), deleteObjectInput)
		if err != nil {
			panic(fmt.Sprintf("Failed to delete object in localstack s3: %s", err.Error()))
		}
	}
}
