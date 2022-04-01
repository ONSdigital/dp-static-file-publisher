package steps

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-net/request"
	s3client "github.com/ONSdigital/dp-s3/v2"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/ONSdigital/log.go/v2/log"
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
	requests              map[string]actualRequest
)

type actualRequest struct {
	body       string
	authHeader string
}

func (c *FilePublisherComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a message to publish the file "([^"]*)" is sent$`, c.aMessageToPublishTheFileIsSent)
	ctx.Step(`^the content of file "([^"]*)" in the public bucket matches the original plain text content$`, c.theContentOfFileInThePublicBucketMatchesTheOriginalPlainTextContent)
	ctx.Step(`^the files API should be informed the file "([^"]*)" has been decrypted$`, c.theFilesAPIShouldBeInformedTheFileHasBeenDecrypted)
	ctx.Step(`^the private bucket still has a encrypted file called "([^"]*)"$`, c.thePrivateBucketStillHasAEncryptedFileCalled)
	ctx.Step(`^the public bucket contains a decrypted file called "([^"]*)"$`, c.thePublicBucketContainsADecryptedFileCalled)
	ctx.Step(`^there is an encrypted single chunk file "([^"]*)" in the private bucket with content:$`, c.thereIsAnEncryptedSingleChunkFileInThePrivateBucketWithContent)
	ctx.Step(`^there is an encryption key for file "([^"]*)" in vault$`, c.thereIsAnEncryptionKeyForFileInVault)
	ctx.Step(`^files API is available$`, c.filesAPIIsAvailable)
	ctx.Step(`^there is an encrypted multi-chunk file "([^"]*)" in the private bucket$`, c.thereIsAnEncryptedMultichunkFileInThePrivateBucket)
}

func (c *FilePublisherComponent) aMessageToPublishTheFileIsSent(file string) error {
	minBrokersHealthy := 1

	ctx := context.Background()

	pub, err := kafka.NewProducer(ctx, &kafka.ProducerConfig{
		KafkaVersion:      &c.config.KafkaVersion,
		MinBrokersHealthy: &minBrokersHealthy,
		Topic:             c.config.StaticFilePublishedTopicV2,
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

	sub, err := kafka.NewConsumerGroup(ctx, &kafka.ConsumerGroupConfig{
		KafkaVersion:      &c.config.KafkaVersion,
		Offset:            nil,
		MinBrokersHealthy: &minBrokersHealthy,
		Topic:             c.config.StaticFilePublishedTopicV2,
		GroupName:         "testing",
		BrokerAddrs:       c.config.KafkaAddr,
	})

	if err != nil {
		log.Error(ctx, "Create Kafka consumer group", err)
	}

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
	assert.Equal(c.ApiFeature, expectedContent, string(b), "Public bucket file content does not match")

	return c.ApiFeature.StepError()
}

func (c *FilePublisherComponent) theFilesAPIShouldBeInformedTheFileHasBeenDecrypted(filename string) error {
	actualRequest := requests[fmt.Sprintf("/files/%s|PATCH", filename)]

	assert.Contains(c.ApiFeature, actualRequest.body, "DECRYPTED")
	assert.NotEqualf(c.ApiFeature, "", actualRequest, "No request body")
	assert.Equal(c.ApiFeature, "Bearer 4424A9F2-B903-40F4-85F1-240107D1AFAF", actualRequest.authHeader, "missing auth token")

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

func (c *FilePublisherComponent) thereIsAnEncryptedSingleChunkFileInThePrivateBucketWithContent(filename string, fileContent *godog.DocString) error {
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
	requests = make(map[string]actualRequest)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		authHeader := r.Header.Get(request.AuthHeaderKey)
		requests[fmt.Sprintf("%s|%s", r.URL.Path, r.Method)] = actualRequest{string(body), authHeader}
		w.WriteHeader(http.StatusOK)
	}))

	os.Setenv("FILES_API_URL", s.URL)

	return nil
}

func (c *FilePublisherComponent) thereIsAnEncryptedMultichunkFileInThePrivateBucket(filename string) error {
	client := s3client.NewClientWithSession(c.config.PrivateBucketName, c.session)

	expectedContentLength = 6 * 1024 * 1024

	content := make([]byte, expectedContentLength)
	rand.Read(content)

	expectedContent = string(content)

	decodeString, _ := hex.DecodeString(encryptionKey)

	for i := 1; i <= 2; i++ {
		var b []byte

		if i == 1 {
			b = content[:(5 * 1024 * 1024)]
		} else {
			b = content[(5 * 1024 * 1024):]
		}

		_, err := client.UploadPartWithPsk(context.Background(), &s3client.UploadPartRequest{
			UploadKey:   filename,
			Type:        "text/plain",
			ChunkNumber: int64(i),
			TotalChunks: 2,
			FileName:    filename,
		},
			b,
			decodeString,
		)

		assert.NoError(c.ApiFeature, err)
	}

	return c.ApiFeature.StepError()
}
