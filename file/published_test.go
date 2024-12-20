package file_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	fileMock "github.com/ONSdigital/dp-static-file-publisher/file/mock"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	errMsg = "s3 is broken"
)

var (
	s3Client   *fileMock.S3ClientMock
	fileClient *fileMock.FilesServiceMock

	nopCloser            = io.NopCloser(strings.NewReader("testing"))
	fileDoesNotExistFunc = func(key string) (bool, error) { return false, nil }
	validGetFunc         = func(key string) (io.ReadCloser, *int64, error) { return nopCloser, nil, nil }
	validUploadFunc      = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
		return &s3manager.UploadOutput{ETag: aws.String("1234567890")}, nil
	}

	errFiles           = errors.New("files error")
	inValidFilesClient = func(ctx context.Context, path string, etag string) error {
		return errFiles
	}
	validFilesClient = func(ctx context.Context, path string, etag string) error {
		return nil
	}
	serviceAuthToken = "Why can't I paste!"
)

func TestHandleFilePublishMessage(t *testing.T) {
	ctx := context.Background()

	s3Client = &fileMock.S3ClientMock{
		FileExistsFunc: fileDoesNotExistFunc,
	}
	fileClient = &fileMock.FilesServiceMock{}

	Convey("Given invalid message content", t, func() {

		msg := MockMessage{}
		msg.Data = []byte("Testing")

		Convey("When the message is handled", func() {
			err := generateMoverCopier().HandleFilePublishMessage(ctx, []kafka.Message{msg})

			Convey("Then a Commit error should be returned", func() {
				So(err, ShouldBeError)
				ensureCommitError(err)
			})
		})
	})

	Convey("Given files in the private bucket could not be read", t, func() {
		s3Client.GetFunc = func(key string) (io.ReadCloser, *int64, error) { return nil, nil, errors.New(errMsg) }
		s3Client.FileExistsFunc = fileDoesNotExistFunc

		Convey("Attempting to get a file from private s3 bucket", func() {
			err := generateMoverCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given files sent the public bucket could not be written", t, func() {
		s3Client.GetFunc = validGetFunc
		s3Client.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			return nil, errors.New(errMsg)
		}

		Convey("Attempting to get a file from private s3 bucket", func() {
			err := generateMoverCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given there already is a file in the public bucket with the provided path", t, func() {
		s3Client.GetFunc = validGetFunc
		s3Client.FileExistsFunc = func(key string) (bool, error) { return true, nil }

		Convey("When the duplicate file is sent for movement", func() {
			err := generateMoverCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithPartMessage(err, "moved file already exists")
			})
		})
	})

	Convey("Given there is a error checking the head of the file in the public bucket", t, func() {
		s3Client.GetFunc = validGetFunc
		s3Client.FileExistsFunc = func(key string) (bool, error) { return false, errors.New(errMsg) }

		Convey("When Head error is returned", func() {
			err := generateMoverCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given the file api returns an error", t, func() {
		s3Client.GetFunc = validGetFunc
		s3Client.FileExistsFunc = fileDoesNotExistFunc
		s3Client.UploadFunc = validUploadFunc
		fileClient.MarkFileMovedFunc = inValidFilesClient

		Convey("When files api returns error (api-client)", func() {
			err := generateMoverCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithPartMessage(err, "files error")
			})
		})
	})

	Convey("Given the file api returns success", t, func() {
		s3Client.GetFunc = validGetFunc
		s3Client.FileExistsFunc = fileDoesNotExistFunc
		s3Client.UploadFunc = validUploadFunc
		fileClient.MarkFileMovedFunc = validFilesClient

		Convey("When files api returns success", func() {
			err := generateMoverCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should not be returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func ensureCommitErrorWithPartMessage(err error, part string) {
	So(err, ShouldBeError)
	So(err.Error(), ShouldContainSubstring, part)
	ensureCommitError(err)
}

func ensureCommitErrorWithMessage(err error, msg string) {
	So(err, ShouldBeError)
	So(err.Error(), ShouldEqual, msg)
	ensureCommitError(err)
}

func ensureCommitError(err error) {
	commiter, ok := err.(kafka.Commiter)

	So(ok, ShouldBeTrue)
	So(commiter.Commit(), ShouldBeTrue)
}

func generateMoverCopier() file.MoverCopier {
	return file.NewMoverCopier(s3Client, s3Client, fileClient)
}

func generateMockMessage() MockMessage {
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

	c, _ := schema.Marshal(file.Published{"test/file.txt", "plain/test", "1234567890", "123"})

	return MockMessage{Data: c}
}

type MockMessage struct {
	Data []byte
}

func (m MockMessage) GetData() []byte             { return m.Data }
func (m MockMessage) Offset() int64               { return 1 }
func (m MockMessage) UpstreamDone() chan struct{} { return nil }
func (m MockMessage) Mark()                       {}
func (m MockMessage) Commit()                     {}
func (m MockMessage) Release()                    {}
func (m MockMessage) CommitAndRelease()           {}
