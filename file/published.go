package file

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//go:generate moq -out mock/s3client.go -pkg mock . S3ClientV2
type S3ClientV2 interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
	FileExists(key string) (bool, error)
}

//go:generate moq -out mock/filesservice.go -pkg mock . FilesService
type FilesService interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	MarkFileDecrypted(ctx context.Context, path string, etag string) error
}

func NewDecrypterCopier(public, private S3ClientV2, vc event.VaultClient, vaultPath string, filesClient FilesService) DecrypterCopier {
	return DecrypterCopier{public, private, vc, vaultPath, filesClient}
}

type DecrypterCopier struct {
	PublicClient  S3ClientV2
	PrivateClient S3ClientV2
	VaultClient   event.VaultClient
	VaultPath     string
	FilesService  FilesService
}

func (d DecrypterCopier) HandleFilePublishMessage(ctx context.Context, msgs []kafka.Message) error {
	logData := log.Data{}
	schema := &avro.Schema{Definition: filePublishAvroSchema}
	fp := Published{}
	log.Info(ctx, fmt.Sprintf("HandleFilePublishMessage (batched) invoked with %d message(s)", len(msgs)))

	for _, msg := range msgs {
		if err := schema.Unmarshal(msg.GetData(), &fp); err != nil {
			return NewCommitError(ctx, err, "Unmarshalling message", logData)
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

		if err := d.notifyFileAPIDecryptionComplete(ctx, uploadResponse, fp, logData); err != nil {
			return err
		}
	}

	return nil
}

func (d DecrypterCopier) notifyFileAPIDecryptionComplete(ctx context.Context, uploadResponse *s3manager.UploadOutput, fp Published, logData log.Data) error {
	err := d.FilesService.MarkFileDecrypted(ctx, fp.Path, *uploadResponse.ETag)
	if err != nil {
		logData["file_path"] = fp.Path
		logData["etag"] = *uploadResponse.ETag
		return NewCommitError(ctx, err, "recieved error from files service", logData)
	}
	return nil
}

func (d DecrypterCopier) decryptAndCopyFile(ctx context.Context, fp Published, encyptionKey []byte, logData log.Data) (*s3manager.UploadOutput, error) {
	reader, _, err := d.PrivateClient.GetWithPSK(fp.Path, encyptionKey)
	if err != nil {
		return nil, NewCommitError(ctx, err, "Reading encrypted file from private s3 bucket", logData)
	}

	uploadResponse, err := d.PublicClient.Upload(&s3manager.UploadInput{
		Key:         &fp.Path,
		ContentType: &fp.Type,
		Body:        reader,
	})
	if err != nil {
		return nil, NewCommitError(ctx, err, "Write decrypted file to public s3 bucket", logData)
	}
	return uploadResponse, nil
}

func (d DecrypterCopier) getEncyptionKey(ctx context.Context, fp Published, logData log.Data) ([]byte, error) {
	keyString, err := d.VaultClient.ReadKey(fmt.Sprintf("%s/%s", d.VaultPath, fp.Path), "key")
	if err != nil {
		return nil, NewCommitError(ctx, err, "Getting encryption key", logData)
	}

	keyBytes, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, NewCommitError(ctx, err, "Decoding encryption key", logData)
	}
	return keyBytes, nil
}

func (d DecrypterCopier) ensurePublicFileDoesNotAlreadyExists(ctx context.Context, fp Published, logData log.Data) error {
	fileExists, err := d.PublicClient.FileExists(fp.Path)
	if err != nil {
		return NewCommitError(ctx, err, "failed to check if file exists", logData)
	}
	if fileExists {
		return NewCommitError(ctx, errors.New("decrypted file already exists"), "File already exists in public bucket", logData)
	}

	return nil
}
