package steps

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	dps3 "github.com/ONSdigital/dp-s3/v3"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

type FilePublished struct {
	Path        string `avro:"path"`
	Type        string `avro:"type"`
	Etag        string `avro:"etag"`
	SizeInBytes string `avro:"sizeInBytes"`
}

var (
	expectedContentLength int
	expectedContent       string
)

func (c *FilePublisherComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a message to publish the file "([^"]*)" is sent$`, c.aMessageToPublishTheFileIsSent)
	ctx.Step(`^the content of file "([^"]*)" in the public bucket matches the original plain text content$`, c.theContentOfFileInThePublicBucketMatchesTheOriginalPlainTextContent)
	ctx.Step(`^the files API should be informed the file "([^"]*)" has been moved$`, c.theFilesAPIShouldBeInformedTheFileHasBeenMoved)
	ctx.Step(`^the private bucket still has a file called "([^"]*)"$`, c.thePrivateBucketStillHasAFileCalled)
	ctx.Step(`^the public bucket contains a moved file called "([^"]*)"$`, c.thePublicBucketContainsAMovedFileCalled)
	ctx.Step(`^there is a single chunk file "([^"]*)" in the private bucket with content:$`, c.thereIsASingleChunkFileInThePrivateBucketWithContent)
	ctx.Step(`^there is a multi-chunk file "([^"]*)" in the private bucket$`, c.thereIsAMultichunkFileInThePrivateBucket)
}

func (c *FilePublisherComponent) aMessageToPublishTheFileIsSent(file string) error {
	minBrokersHealthy := 1

	ctx := context.Background()

	pub, err := kafka.NewProducer(ctx, &kafka.ProducerConfig{
		KafkaVersion:      &c.config.KafkaVersion,
		MinBrokersHealthy: &minBrokersHealthy,
		Topic:             c.config.StaticFilePublishedTopic,
		BrokerAddrs:       c.config.KafkaAddr,
	})

	if err != nil {
		log.Error(ctx, "New Kafka producer", err)
	}

	schema := &avro.Schema{
		Definition: `{
					"type": "record",
					"name": "file-published",
					"fields": [
					  {"name": "path", "type": "string"},
					  {"name": "etag", "type": "string"},
					  {"name": "type", "type": "string"},
					  {"name": "sizeInBytes", "type": "string"}
					]
				  }`,
	}

	msg := &FilePublished{
		Path:        file,
		Etag:        "TODO get etag when saving file",
		Type:        "text/plain",
		SizeInBytes: "TODO get size of file from file",
	}

	err = pub.Send(schema, msg)

	if err != nil {
		log.Error(ctx, "Publish send", err)
	}

	c.Initialiser()

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) theContentOfFileInThePublicBucketMatchesTheOriginalPlainTextContent(filename string) error {
	client := dps3.NewClientWithConfig(c.config.PublicBucketName, *c.AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})
	result, _, err := client.Get(context.Background(), filename)

	b, _ := io.ReadAll(result)

	assert.NoError(c.ApiFeature, err)
	assert.Equal(c.ApiFeature, expectedContent, string(b), "Public bucket file content does not match")

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) theFilesAPIShouldBeInformedTheFileHasBeenMoved(filename string) error {
	if len(c.request) > 0 {
		_, ok := c.request[filename]
		assert.True(c.ApiFeature, ok, fmt.Sprintf("expecting files-api move to be invoked with %s", filename))
	}
	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thePrivateBucketStillHasAFileCalled(filename string) error {
	client := dps3.NewClientWithConfig(c.config.PrivateBucketName, *c.AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})
	result, err := client.Head(context.Background(), filename)

	assert.Equal(c.ApiFeature, expectedContentLength, int(*result.ContentLength))
	assert.Equal(c.ApiFeature, "text/plain", *result.ContentType)

	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thePublicBucketContainsAMovedFileCalled(filename string) error {
	client := dps3.NewClientWithConfig(c.config.PublicBucketName, *c.AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})

	result, err := client.Head(context.Background(), filename)

	assert.NoError(c.ApiFeature, err)

	if err == nil {
		assert.Equal(c.ApiFeature, expectedContentLength, int(*result.ContentLength))
		assert.Equal(c.ApiFeature, "text/plain", *result.ContentType)
	}

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thereIsASingleChunkFileInThePrivateBucketWithContent(filename string, fileContent *godog.DocString) error {
	client := dps3.NewClientWithConfig(c.config.PrivateBucketName, *c.AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})

	expectedContentLength = len(fileContent.Content)
	expectedContent = fileContent.Content

	_, err := client.UploadPart(context.Background(), &dps3.UploadPartRequest{
		UploadKey:   filename,
		Type:        "text/plain",
		ChunkNumber: 1,
		TotalChunks: 1,
		FileName:    filename,
	},
		[]byte(fileContent.Content),
	)

	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thereIsAMultichunkFileInThePrivateBucket(filename string) error {
	client := dps3.NewClientWithConfig(c.config.PrivateBucketName, *c.AWSConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(localStackHost)
		options.UsePathStyle = true
	})

	expectedContentLength = 6 * 1024 * 1024

	content := make([]byte, expectedContentLength)
	rand.Read(content)

	expectedContent = string(content)

	for i := 1; i <= 2; i++ {
		var b []byte

		if i == 1 {
			b = content[:(5 * 1024 * 1024)]
		} else {
			b = content[(5 * 1024 * 1024):]
		}

		_, err := client.UploadPart(context.Background(), &dps3.UploadPartRequest{
			UploadKey:   filename,
			Type:        "text/plain",
			ChunkNumber: int32(i),
			TotalChunks: 2,
			FileName:    filename,
		},
			b,
		)

		assert.NoError(c.ApiFeature, err)
	}

	return c.ApiFeature.StepError()
}
