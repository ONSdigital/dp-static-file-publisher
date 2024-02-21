package steps

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	s3client "github.com/ONSdigital/dp-s3/v2"
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
	client := s3client.NewClientWithSession(c.config.PublicBucketName, c.session)
	result, _, err := client.Get(filename)

	b, _ := io.ReadAll(result)

	assert.NoError(c.ApiFeature, err)
	assert.Equal(c.ApiFeature, expectedContent, string(b), "Public bucket file content does not match")

	return c.ApiFeature.StepError()
}
