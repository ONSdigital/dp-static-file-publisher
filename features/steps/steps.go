package steps

import (
	"context"
	"fmt"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/ONSdigital/dp-static-file-publisher/config"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cucumber/godog"
	"sync"
)

type FilePublished struct {
	Path        string `avro:"path"`
	Type        string `avro:"type"`
	Etag        string `avro:"etag"`
	SizeInBytes string `avro:"sizeInBytes"`
}

func (c *FilePublisherComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a message to publish the file "([^"]*)" is sent$`, c.aMessageToPublishTheFileIsSent)
	ctx.Step(`^the content of file "([^"]*)" in the public bucket matches the original plain text content$`, c.theContentOfFileInThePublicBucketMatchesTheOriginalPlainTextContent)
	ctx.Step(`^the files API should be informed the file has been decrypted$`, c.theFilesAPIShouldBeInformedTheFileHasBeenDecrypted)
	ctx.Step(`^the private bucket still has a encrypted file called "([^"]*)"$`, c.thePrivateBucketStillHasAEncryptedFileCalled)
	ctx.Step(`^the public bucket contains a decrypted file called "([^"]*)"$`, c.thePublicBucketContainsADecryptedFileCalled)
	ctx.Step(`^there is a encrypted file "([^"]*)" in the private bucket$`, c.thereIsAEncryptedSingleChunkFileInThePrivateBucketWithContent)
	ctx.Step(`^there is an encryption key for file "([^"]*)" in vault$`, c.thereIsAnEncryptionKeyForFileInVault)
}

func (c *FilePublisherComponent) aMessageToPublishTheFileIsSent(file string) error {
	cfg, _ := config.Get()

	minBrokersHealthy := 1
	ctx := context.Background()
	pub, _ := kafka.NewProducer(ctx, &kafka.ProducerConfig{
		KafkaVersion:      &cfg.KafkaVersion,
		MinBrokersHealthy: &minBrokersHealthy,
		Topic:             cfg.StaticFilePublishedTopicV2,
		BrokerAddrs:       cfg.KafkaAddr,
	})

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

	pub.Send(schema, msg)

	c.Initialiser()

	sub, _ := kafka.NewConsumerGroup(ctx, &kafka.ConsumerGroupConfig{
		KafkaVersion:      &cfg.KafkaVersion,
		Offset:            nil,
		MinBrokersHealthy: &minBrokersHealthy,
		Topic:             cfg.StaticFilePublishedTopicV2,
		GroupName:         "testing",
		BrokerAddrs:       cfg.KafkaAddr,
	})

	sub.Start()
	wg := sync.WaitGroup{}
	wg.Add(1)
	sub.RegisterHandler(ctx, func(ctx context.Context, workerID int, msg kafka.Message) error {
		wg.Done()

		return nil
	})

	wg.Wait()

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) theContentOfFileInThePublicBucketMatchesTheOriginalPlainTextContent(arg1 string) error {
	return godog.ErrPending
}

func (c *FilePublisherComponent) theFilesAPIShouldBeInformedTheFileHasBeenDecrypted() error {
	return godog.ErrPending
}

func (c *FilePublisherComponent) thePrivateBucketStillHasAEncryptedFileCalled(arg1 string) error {
	return godog.ErrPending
}

func (c *FilePublisherComponent) thePublicBucketContainsADecryptedFileCalled(filename string) error {
	return godog.ErrPending
}

func (c *FilePublisherComponent) thereIsAEncryptedSingleChunkFileInThePrivateBucketWithContent(filename string, fileContent *godog.DocString) error {
	cfg, _ := config.Get()

	s, _ := session.NewSession(&aws.Config{
		Endpoint:         aws.String(localStackHost),
		Region:           aws.String(cfg.AwsRegion),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("test", "test", "")})

	client := s3client.NewClientWithSession(cfg.PrivateBucketName, s)

	client.UploadPartWithPsk(context.Background(), &s3client.UploadPartRequest{
		UploadKey:   filename,
		Type:        "text/plain",
		ChunkNumber: 1,
		TotalChunks: 1,
		FileName:    filename,
	},
		[]byte(fileContent.Content),
		[]byte(encryptionKey),
	)

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thereIsAnEncryptionKeyForFileInVault(filename string) error {
	cfg, _ := config.Get()
	v, _ := vault.CreateClient(cfg.VaultToken, cfg.VaultAddress, 5)

	v.WriteKey(fmt.Sprintf("%s/%s", cfg.VaultPath, filename), "key", encryptionKey)

	return c.ApiFeature.StepError()
}
