package file

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"io"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:generate moq -out mock/s3client.go -pkg mock . S3Client
type S3Client interface {
	BucketName() string
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	FileExists(ctx context.Context, key string) (bool, error)
	Get(ctx context.Context, key string) (io.ReadCloser, *int64, error)
	Upload(ctx context.Context, input *s3.PutObjectInput, options ...func(*manager.Uploader)) (*manager.UploadOutput, error)
	Config() aws.Config
}

//go:generate moq -out mock/filesservice.go -pkg mock . FilesService
type FilesService interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	MarkFileMoved(ctx context.Context, path string, etag string) error
}

func NewMoverCopier(public, private S3Client, filesClient FilesService) MoverCopier {
	return MoverCopier{public, private, filesClient}
}

type MoverCopier struct {
	PublicClient  S3Client
	PrivateClient S3Client
	FilesService  FilesService
}

func (d MoverCopier) HandleFilePublishMessage(ctx context.Context, msgs []kafka.Message) error {
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

		uploadResponse, err := d.moveAndCopyFile(ctx, fp, logData)
		if err != nil {
			return err
		}

		if err := d.notifyFileAPIMovementComplete(ctx, uploadResponse, fp, logData); err != nil {
			return err
		}
	}

	return nil
}

func (d MoverCopier) notifyFileAPIMovementComplete(ctx context.Context, uploadResponse *manager.UploadOutput, fp Published, logData log.Data) error {
	err := d.FilesService.MarkFileMoved(ctx, fp.Path, *uploadResponse.ETag)
	if err != nil {
		logData["file_path"] = fp.Path
		logData["etag"] = *uploadResponse.ETag
		return NewCommitError(ctx, err, "recieved error from files service", logData)
	}
	return nil
}

func (d MoverCopier) moveAndCopyFile(ctx context.Context, fp Published, logData log.Data) (*manager.UploadOutput, error) {
	reader, _, err := d.PrivateClient.Get(ctx, fp.Path)
	if err != nil {
		return nil, NewCommitError(ctx, err, "Reading file from private s3 bucket", logData)
	}

	uploadResponse, err := d.PublicClient.Upload(ctx, &s3.PutObjectInput{
		Key:         &fp.Path,
		ContentType: &fp.Type,
		Body:        reader,
	})
	if err != nil {
		return nil, NewCommitError(ctx, err, "Write moved file to public s3 bucket", logData)
	}
	return uploadResponse, nil
}

func (d MoverCopier) ensurePublicFileDoesNotAlreadyExists(ctx context.Context, fp Published, logData log.Data) error {
	fileExists, err := d.PublicClient.FileExists(ctx, fp.Path)
	if err != nil {
		return NewCommitError(ctx, err, "failed to check if file exists", logData)
	}
	if fileExists {
		return NewCommitError(ctx, errors.New("moved file already exists"), "File already exists in public bucket", logData)
	}

	return nil
}
