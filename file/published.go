package file

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//go:generate moq -out mock/s3client.go -pkg mock . S3ClientV2
type S3ClientV2 interface {
	BucketName() string
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	FileExists(key string) (bool, error)
	Get(key string) (io.ReadCloser, *int64, error)
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	Session() *session.Session
}

//go:generate moq -out mock/filesservice.go -pkg mock . FilesService
type FilesService interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	MarkFileMoved(ctx context.Context, path string, etag string) error
}

func NewMoverCopier(public, private S3ClientV2, filesClient FilesService) MoverCopier {
	return MoverCopier{public, private, filesClient}
}

type MoverCopier struct {
	PublicClient  S3ClientV2
	PrivateClient S3ClientV2
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

func (d MoverCopier) notifyFileAPIMovementComplete(ctx context.Context, uploadResponse *s3manager.UploadOutput, fp Published, logData log.Data) error {
	err := d.FilesService.MarkFileMoved(ctx, fp.Path, *uploadResponse.ETag)
	if err != nil {
		logData["file_path"] = fp.Path
		logData["etag"] = *uploadResponse.ETag
		return NewCommitError(ctx, err, "recieved error from files service", logData)
	}
	return nil
}

func (d MoverCopier) moveAndCopyFile(ctx context.Context, fp Published, logData log.Data) (*s3manager.UploadOutput, error) {
	reader, _, err := d.PrivateClient.Get(fp.Path)
	if err != nil {
		return nil, NewCommitError(ctx, err, "Reading file from private s3 bucket", logData)
	}

	uploadResponse, err := d.PublicClient.Upload(&s3manager.UploadInput{
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
	fileExists, err := d.PublicClient.FileExists(fp.Path)
	if err != nil {
		return NewCommitError(ctx, err, "failed to check if file exists", logData)
	}
	if fileExists {
		return NewCommitError(ctx, errors.New("moved file already exists"), "File already exists in public bucket", logData)
	}

	return nil
}
