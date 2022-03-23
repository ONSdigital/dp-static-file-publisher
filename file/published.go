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
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"net/http"
)

//go:generate moq -out mock/s3client.go -pkg mock . S3ClientV2
type S3ClientV2 interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
	FileExists(key string) (bool, error)
}

func NewDecrypterCopier(public, private S3ClientV2, vc event.VaultClient, vaultPath, filesAPIURL string) DecrypterCopier {
	return DecrypterCopier{public, private, vc, vaultPath, filesAPIURL}
}

type DecrypterCopier struct {
	PublicClient  S3ClientV2
	PrivateClient S3ClientV2
	VaultClient   event.VaultClient
	VaultPath     string
	FilesAPIURL   string
}

func (d DecrypterCopier) HandleFilePublishMessage(ctx context.Context, workerID int, msg kafka.Message) error {
	logData := log.Data{}
	schema := &avro.Schema{Definition: filePublishAvroSchema}
	fp := Published{}

	if err := schema.Unmarshal(msg.GetData(), &fp); err != nil {
		return NewNoCommitError(ctx, err, "Unmarshalling message", logData)
	}

	logData["file-published-message"] = fp

	if err := d.ensurePublicFileDoesNotAlreadyExists(ctx, fp, logData); err != nil {
		return err
	}

	encyptionKey, err := d.getEncyptionKey(ctx, fp, logData)
	if err != nil {
		return err
	}

	uploadResponse, err := d.decryptAndCopyFile(ctx, fp, encyptionKey, logData)
	if err != nil {
		return err
	}

	if err := d.notifyFileApiDecryptionComplete(ctx, uploadResponse, fp, logData); err != nil {
		return err
	}

	return nil
}

func (d DecrypterCopier) notifyFileApiDecryptionComplete(ctx context.Context, uploadResponse *s3manager.UploadOutput, fp Published, logData log.Data) error {
	const stateDecrypted = "DECRYPTED"
	type FilesAPIRequestBody struct {
		Etag  string
		State string
	}

	requestBody := FilesAPIRequestBody{
		Etag:  *uploadResponse.ETag,
		State: stateDecrypted,
	}

	filesAPIPath := fmt.Sprintf("%s/files/%s", d.FilesAPIURL, fp.Path)
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPatch, filesAPIPath, bytes.NewReader(body))

	response, err := dphttp.DefaultClient.Do(ctx, req)
	if err != nil {
		logData["request"] = req
		logData["request_body"] = string(body)
		return NewNoCommitError(ctx, err, "could not send HTTP request", logData)
	}

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		logData["response"] = string(body)

		switch response.StatusCode {
		case http.StatusNotFound:
			err = errors.New("file not found on dp-files-api")
		case http.StatusBadRequest:
			err = errors.New("invalid request to dp-files-api")
		case http.StatusInternalServerError:
			err = errors.New("error received from dp-files-api")
		case http.StatusUnauthorized:
			err = errors.New("unauthorized request to dp-files-api")
		case http.StatusForbidden:
			err = errors.New("forbidden request to dp-files-api")
		default:
			err = errors.New("unexpected response from dp-files-api")
		}
		log.Error(ctx, "failed response from dp-files-api", err, logData)

		return NewNoCommitError(ctx, err, "HTTP Error", logData)

	}
	return nil
}

func (d DecrypterCopier) decryptAndCopyFile(ctx context.Context, fp Published, encyptionKey []byte, logData log.Data) (*s3manager.UploadOutput, error) {
	reader, _, err := d.PrivateClient.GetWithPSK(fp.Path, encyptionKey)
	if err != nil {
		return nil, NewNoCommitError(ctx, err, "Reading encrypted file from private s3 bucket", logData)
	}

	uploadResponse, err := d.PublicClient.Upload(&s3manager.UploadInput{
		Key:         &fp.Path,
		ContentType: &fp.Type,
		Body:        reader,
	})
	if err != nil {
		return nil, NewNoCommitError(ctx, err, "Write decrypted file to public s3 bucket", logData)
	}
	return uploadResponse, nil
}

func (d DecrypterCopier) getEncyptionKey(ctx context.Context, fp Published, logData log.Data) ([]byte, error) {
	keyString, err := d.VaultClient.ReadKey(fmt.Sprintf("%s/%s", d.VaultPath, fp.Path), "key")
	if err != nil {
		return nil, NewNoCommitError(ctx, err, "Getting encryption key", logData)
	}

	keyBytes, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, NewNoCommitError(ctx, err, "Decoding encryption key", logData)
	}
	return keyBytes, nil
}

func (d DecrypterCopier) ensurePublicFileDoesNotAlreadyExists(ctx context.Context, fp Published, logData log.Data) error {
	fileExists, err := d.PublicClient.FileExists(fp.Path)
	if err != nil {
		return NewNoCommitError(ctx, err, "failed to check if file exists", logData)
	}
	if fileExists {
		return NewNoCommitError(ctx, errors.New("decrypted file already exists"), "File already exists in public bucket", logData)
	}

	return nil
}
