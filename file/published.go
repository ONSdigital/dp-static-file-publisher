package file

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"net/http"
)

const (
	stateDecrypted = "DECRYPTED"
)

type Published struct {
	Path        string `avro:"path"`
	Type        string `avro:"type"`
	Etag        string `avro:"etag"`
	SizeInBytes string `avro:"sizeInBytes"`
}

//go:generate moq -out mock/s3client.go -pkg mock . S3ClientV2
type S3ClientV2 interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
	FileExists(key string) (bool, error)
}

type FilesAPIRequestBody struct {
	Etag  string
	State string
}

type DecrypterCopier struct {
	PublicClient  S3ClientV2
	PrivateClient S3ClientV2
	VaultClient   event.VaultClient
	VaultPath     string
	FilesAPIURL   string
}

type NoCommitError struct {
	err error
}

func (n NoCommitError) Commit() bool {
	return false
}

func (n NoCommitError) Error() string {
	return n.err.Error()
}

func (d DecrypterCopier) HandleFilePublishMessage(ctx context.Context, workerID int, msg kafka.Message) error {
	logData := log.Data{}
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
	fp := Published{}

	err := schema.Unmarshal(msg.GetData(), &fp)
	if err != nil {
		log.Error(ctx, "Unmarshalling message", err, logData)
		return NoCommitError{err}
	}

	logData["file-published-message"] = fp

	fileExists, err := d.PublicClient.FileExists(fp.Path)
	if err != nil {
		log.Error(ctx, "failed to check if file exists", err, logData)
		return NoCommitError{err}
	}
	if fileExists {
		err = errors.New("decrypted file already exists")
		log.Error(ctx, "File already exists in public bucket", err, logData)
		return NoCommitError{err}
	}

	encryptionKey, err := d.VaultClient.ReadKey(fmt.Sprintf("%s/%s", d.VaultPath, fp.Path), "key")

	if err != nil {
		log.Error(ctx, "Getting encryption key", err, logData)
		return NoCommitError{err}
	}

	decodeString, err := hex.DecodeString(encryptionKey)

	if err != nil {
		log.Error(ctx, "Decoding encryption key", err, logData)
		return NoCommitError{err}
	}

	reader, _, err := d.PrivateClient.GetWithPSK(fp.Path, decodeString)

	if err != nil {
		log.Error(ctx, "Reading encrypted file from private s3 bucket", err, logData)
		return NoCommitError{err}
	}

	uploadResponse, err := d.PublicClient.Upload(&s3manager.UploadInput{
		Key:         &fp.Path,
		ContentType: &fp.Type,
		Body:        reader,
	})

	if err != nil {
		log.Error(ctx, "Write decrypted file to public s3 bucket", err, logData)
		return NoCommitError{err}
	}

	requestBody := FilesAPIRequestBody{
		Etag:  *uploadResponse.ETag,
		State: stateDecrypted,
	}

	filesAPIPath := fmt.Sprintf("%s/files/%s", d.FilesAPIURL, fp.Path)
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPatch, filesAPIPath, bytes.NewReader(body))

	hc := http.Client{}
	_, err = hc.Do(req)

	if err != nil {
		log.Error(ctx, "FILES API REQUEST ERROR", err)
	}

	log.Info(ctx, "Finished request to files API")

	return nil
}
