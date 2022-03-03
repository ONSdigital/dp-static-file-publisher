package steps

import (
	"context"
	"encoding/hex"
	"fmt"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	s3client "github.com/ONSdigital/dp-s3/v2"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"time"
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
	requests              map[string]string
)

func (c *FilePublisherComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a message to publish the file "([^"]*)" is sent$`, c.aMessageToPublishTheFileIsSent)
	ctx.Step(`^the content of file "([^"]*)" in the public bucket matches the original plain text content$`, c.theContentOfFileInThePublicBucketMatchesTheOriginalPlainTextContent)
	ctx.Step(`^the files API should be informed the file has been decrypted$`, c.theFilesAPIShouldBeInformedTheFileHasBeenDecrypted)
	ctx.Step(`^the private bucket still has a encrypted file called "([^"]*)"$`, c.thePrivateBucketStillHasAEncryptedFileCalled)
	ctx.Step(`^the public bucket contains a decrypted file called "([^"]*)"$`, c.thePublicBucketContainsADecryptedFileCalled)
	ctx.Step(`^there is a encrypted single chunk file "([^"]*)" in the private bucket with content:$`, c.thereIsAEncryptedSingleChunkFileInThePrivateBucketWithContent)
	ctx.Step(`^there is an encryption key for file "([^"]*)" in vault$`, c.thereIsAnEncryptionKeyForFileInVault)
	ctx.Step(`^files API is available$`, c.filesAPIIsAvailable)
}

func (c *FilePublisherComponent) aMessageToPublishTheFileIsSent(file string) error {
	minBrokersHealthy := 1
	ctx := context.Background()
	pub, _ := kafka.NewProducer(ctx, &kafka.ProducerConfig{
		KafkaVersion:      &c.config.KafkaVersion,
		MinBrokersHealthy: &minBrokersHealthy,
		Topic:             c.config.StaticFilePublishedTopicV2,
		BrokerAddrs:       c.config.KafkaAddr,
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
		KafkaVersion:      &c.config.KafkaVersion,
		Offset:            nil,
		MinBrokersHealthy: &minBrokersHealthy,
		Topic:             c.config.StaticFilePublishedTopicV2,
		GroupName:         "testing",
		BrokerAddrs:       c.config.KafkaAddr,
	})

	sub.Start()
	wg := sync.WaitGroup{}
	wg.Add(1)
	sub.RegisterHandler(ctx, func(ctx context.Context, workerID int, msg kafka.Message) error {
		time.Sleep(2 * time.Second)
		wg.Done()

		return nil
	})

	wg.Wait()
	sub.Stop()

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) theContentOfFileInThePublicBucketMatchesTheOriginalPlainTextContent(filename string) error {
	client := s3client.NewClientWithSession(c.config.PublicBucketName, c.session)
	result, _, err := client.Get(filename)

	b, _ := io.ReadAll(result)

	assert.NoError(c.ApiFeature, err)
	assert.Equal(c.ApiFeature, expectedContent, string(b))

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) theFilesAPIShouldBeInformedTheFileHasBeenDecrypted() error {
	body := requests["/files/data/single-chunk.txt|PATCH"]

	assert.Contains(c.ApiFeature, body, "DECRYPTED")
	assert.NotEqualf(c.ApiFeature, "", body, "No request body")

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thePrivateBucketStillHasAEncryptedFileCalled(filename string) error {
	client := s3client.NewClientWithSession(c.config.PrivateBucketName, c.session)
	result, err := client.Head(filename)

	assert.Equal(c.ApiFeature, expectedContentLength, int(*result.ContentLength))
	assert.Equal(c.ApiFeature, "text/plain", *result.ContentType)

	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thePublicBucketContainsADecryptedFileCalled(filename string) error {
	client := s3client.NewClientWithSession(c.config.PublicBucketName, c.session)
	result, err := client.Head(filename)

	assert.NoError(c.ApiFeature, err)

	if err == nil {
		assert.Equal(c.ApiFeature, expectedContentLength, int(*result.ContentLength))
		assert.Equal(c.ApiFeature, "text/plain", *result.ContentType)
	}

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thereIsAEncryptedSingleChunkFileInThePrivateBucketWithContent(filename string, fileContent *godog.DocString) error {
	client := s3client.NewClientWithSession(c.config.PrivateBucketName, c.session)

	expectedContentLength = len(fileContent.Content)
	expectedContent = fileContent.Content

	decodeString, _ := hex.DecodeString(encryptionKey)
	_, err := client.UploadPartWithPsk(context.Background(), &s3client.UploadPartRequest{
		UploadKey:   filename,
		Type:        "text/plain",
		ChunkNumber: 1,
		TotalChunks: 1,
		FileName:    filename,
	},
		[]byte(fileContent.Content),
		decodeString,
	)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) thereIsAnEncryptionKeyForFileInVault(filename string) error {
	v, _ := vault.CreateClient(c.config.VaultToken, c.config.VaultAddress, 5)

	err := v.WriteKey(fmt.Sprintf("%s/%s", c.config.VaultPath, filename), "key", encryptionKey)
	assert.NoError(c.ApiFeature, err)

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) filesAPIIsAvailable() error {
	requests = make(map[string]string)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		requests[fmt.Sprintf("%s|%s", r.URL.Path, r.Method)] = string(body)
		w.WriteHeader(http.StatusOK)
	}))

	os.Setenv("FILES_API_URL", s.URL)

	return nil
}
