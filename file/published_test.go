package file_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	eventMock "github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	fileMock "github.com/ONSdigital/dp-static-file-publisher/file/mock"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	errMsg    = "s3 is broken"
	vaultPath = "testing/secret"
)

var (
	s3Client    *fileMock.S3ClientV2Mock
	vaultClient *eventMock.VaultClientMock
	fileClient  *fileMock.FilesServiceMock

	nopCloser            = io.NopCloser(strings.NewReader("testing"))
	validVaultReadFunc   = func(path string, key string) (string, error) { return "1234567890123456", nil }
	fileDoesNotExistFunc = func(key string) (bool, error) { return false, nil }
	validGetWithPSKFunc  = func(key string, psk []byte) (io.ReadCloser, *int64, error) { return nopCloser, nil, nil }
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

	s3Client = &fileMock.S3ClientV2Mock{
		FileExistsFunc: fileDoesNotExistFunc,
	}
	vaultClient = &eventMock.VaultClientMock{}
	fileClient = &fileMock.FilesServiceMock{}

	Convey("Given invalid message content", t, func() {

		msg := MockMessage{}
		msg.Data = []byte("Testing")

		Convey("When the message is handled", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{msg})

			Convey("Then a Commit error should be returned", func() {
				So(err, ShouldBeError)
				ensureCommitError(err)
			})
		})
	})

	Convey("Given there is a read error on the Vault key", t, func() {
		vaultClient.ReadKeyFunc = func(path string, key string) (string, error) { return "", errors.New("broken") }

		Convey("When the message is handled", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				So(err, ShouldBeError)
				ensureCommitError(err)
			})
		})
	})

	Convey("Given there a specific vault read error", t, func() {
		var errorScenarios = []struct {
			err     error
			errName string
		}{
			{vault.ErrKeyNotFound, "ErrKeyNotFound"},
			{vault.ErrVersionNotFound, "ErrVersionNotFound"},
			{vault.ErrMetadataNotFound, "ErrMetadataNotFound"},
			{vault.ErrDataNotFound, "ErrDataNotFound"},
			{vault.ErrVersionInvalid, "ErrVersionInvalid"},
		}

		for _, scenario := range errorScenarios {
			Convey("And the error is "+scenario.errName, func() {
				vaultClient.ReadKeyFunc = func(path string, key string) (string, error) { return "", scenario.err }

				Convey("When the message is handled", func() {
					err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

					Convey("Then a Commit error should be returned wrapping the original error", func() {
						ensureCommitErrorWithMessage(err, scenario.err.Error())
					})
				})
			})
		}
	})

	Convey("Given the encryption key from vault cannot be parsed to a byte array", t, func() {
		vaultClient.ReadKeyFunc = func(path string, key string) (string, error) { return "abcdefgh", nil }

		Convey("When the message is handled", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithPartMessage(err, "encoding/hex:")
			})
		})
	})

	Convey("Given files in the private bucket could not be read", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) { return nil, nil, errors.New(errMsg) }
		s3Client.FileExistsFunc = fileDoesNotExistFunc

		Convey("Attempting to get a file from to  private s3 bucket", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given files sent the public bucket could not be written", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			return nil, errors.New(errMsg)
		}

		Convey("Attempting to get a file from to  private s3 bucket", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given there already is a file in the public bucket with the provided path", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.FileExistsFunc = func(key string) (bool, error) { return true, nil }

		Convey("When the duplicate file is sent for decryption", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithPartMessage(err, "decrypted file already exists")
			})
		})
	})

	Convey("Given there is a error checking the head of the file in the public bucket", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.FileExistsFunc = func(key string) (bool, error) { return false, errors.New(errMsg) }

		Convey("When Head error is returned", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given the file api returns an error", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.FileExistsFunc = fileDoesNotExistFunc
		s3Client.UploadFunc = validUploadFunc
		fileClient.MarkFileDecryptedFunc = inValidFilesClient

		Convey("When files api returns error (api-client)", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

			Convey("Then a Commit error should be returned", func() {
				ensureCommitErrorWithPartMessage(err, "files error")
			})
		})
	})

	Convey("Given the file api returns success", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.FileExistsFunc = fileDoesNotExistFunc
		s3Client.UploadFunc = validUploadFunc
		fileClient.MarkFileDecryptedFunc = validFilesClient

		Convey("When files api returns success", func() {
			err := generateDecrypterCopier().HandleFilePublishMessage(ctx, []kafka.Message{generateMockMessage()})

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

func generateDecrypterCopier() file.DecrypterCopier {
	return file.NewDecrypterCopier(s3Client, s3Client, vaultClient, vaultPath, fileClient)
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
func (m MockMessage) Context() context.Context    { return nil }
func (m MockMessage) GetHeader(key string) string { return "" }
