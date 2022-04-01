package file_test

import (
	"context"
	"errors"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	eventMock "github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/ONSdigital/dp-static-file-publisher/file"
	fileMock "github.com/ONSdigital/dp-static-file-publisher/file/mock"
	vault "github.com/ONSdigital/dp-vault"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	errMsg    = "s3 is broken"
	vaultPath = "testing/secret"
)

var (
	s3Client    *fileMock.S3ClientV2Mock
	vaultClient *eventMock.VaultClientMock

	nopCloser            = io.NopCloser(strings.NewReader("testing"))
	validVaultReadFunc   = func(path string, key string) (string, error) { return "1234567890123456", nil }
	fileDoesNotExistFunc = func(key string) (bool, error) { return false, nil }
	validGetWithPSKFunc  = func(key string, psk []byte) (io.ReadCloser, *int64, error) { return nopCloser, nil, nil }
	validUploadFunc      = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
		return &s3manager.UploadOutput{ETag: aws.String("1234567890")}, nil
	}
	serviceAuthToken = "Why can't I paste!"
)

func TestHandleFilePublishMessage(t *testing.T) {
	ctx := context.Background()

	s3Client = &fileMock.S3ClientV2Mock{
		FileExistsFunc: fileDoesNotExistFunc,
	}
	vaultClient = &eventMock.VaultClientMock{}

	Convey("Given invalid message content", t, func() {

		msg := MockMessage{}
		msg.Data = []byte("Testing")

		Convey("When the message is handled", func() {
			err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, msg)

			Convey("Then a No Commit error should be returned", func() {
				So(err, ShouldBeError)
				ensureNoCommitError(err)
			})
		})
	})

	Convey("Given there is a read error on the Vault key", t, func() {
		vaultClient.ReadKeyFunc = func(path string, key string) (string, error) { return "", errors.New("broken") }

		Convey("When the message is handled", func() {
			err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, generateMockMessage())

			Convey("Then a No Commit error should be returned", func() {
				So(err, ShouldBeError)
				ensureNoCommitError(err)
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
					err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, generateMockMessage())

					Convey("Then a No Commit error should be returned wrapping the original error", func() {
						ensureNoCommitErrorWithMessage(err, scenario.err.Error())
					})
				})
			})
		}
	})

	Convey("Given the encryption key from vault cannot be parsed to a byte array", t, func() {
		vaultClient.ReadKeyFunc = func(path string, key string) (string, error) { return "abcdefgh", nil }

		Convey("When the message is handled", func() {
			err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, generateMockMessage())

			Convey("Then a No Commit error should be returned", func() {
				ensureNoCommitErrorWithPartMessage(err, "encoding/hex:")
			})
		})
	})

	Convey("Given files in the private bucket could not be read", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) { return nil, nil, errors.New(errMsg) }
		s3Client.FileExistsFunc = fileDoesNotExistFunc

		Convey("Attempting to get a file from to  private s3 bucket", func() {
			err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, generateMockMessage())

			Convey("Then a No Commit error should be returned", func() {
				ensureNoCommitErrorWithMessage(err, errMsg)
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
			err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, generateMockMessage())

			Convey("Then a No Commit error should be returned", func() {
				ensureNoCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given there already is a file in the public bucket with the provided path", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.FileExistsFunc = func(key string) (bool, error) { return true, nil }

		Convey("When the duplicate file is sent for decryption", func() {
			err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, generateMockMessage())

			Convey("Then a No Commit error should be returned", func() {
				ensureNoCommitErrorWithPartMessage(err, "decrypted file already exists")
			})
		})
	})

	Convey("Given there is a error checking the head of the file in the public bucket", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.FileExistsFunc = func(key string) (bool, error) { return false, errors.New(errMsg) }

		Convey("When Head error is returned", func() {
			err := generateDecrypterCopier("does not matter").HandleFilePublishMessage(ctx, 1, generateMockMessage())

			Convey("Then a No Commit error should be returned", func() {
				ensureNoCommitErrorWithMessage(err, errMsg)
			})
		})
	})

	Convey("Given the file api hostname is invalid", t, func() {
		vaultClient.ReadKeyFunc = validVaultReadFunc

		s3Client.GetWithPSKFunc = validGetWithPSKFunc
		s3Client.FileExistsFunc = fileDoesNotExistFunc
		s3Client.UploadFunc = validUploadFunc
		Convey("When the hostname is invalid", func() {
			err := generateDecrypterCopier("invalid").HandleFilePublishMessage(ctx, 1, generateMockMessage())

			Convey("Then a No Commit error should be returned", func() {
				ensureNoCommitErrorWithPartMessage(err, "unsupported protocol scheme")
			})
		})
	})

	Convey("Given the files API returns an error", t, func() {
		var errorScenarios = []struct {
			description string
			status      int
			errorString string
			willCommit  Assertion
		}{
			{
				"Given the file does not exists on Files API",
				http.StatusNotFound,
				"file not found on dp-files-api",
				ShouldBeFalse,
			},
			{
				"Given the file is in an unexpected state",
				http.StatusBadRequest,
				"invalid request to dp-files-api",
				ShouldBeFalse,
			},
			{
				"Given the file API has an internal error",
				http.StatusInternalServerError,
				"error received from dp-files-api",
				ShouldBeFalse,
			},
			{
				"Given we are unauthorized to contact dp-files-api",
				http.StatusUnauthorized,
				"unauthorized request to dp-files-api",
				ShouldBeFalse,
			},
			{
				"Given we are forbidden to contact dp-files-api",
				http.StatusForbidden,
				"forbidden request to dp-files-api",
				ShouldBeFalse,
			},
			{
				"Given dp-files-api returns an unexpect http response",
				http.StatusTeapot,
				"unexpected response from dp-files-api",
				ShouldBeFalse,
			},
		}

		for _, errorScenario := range errorScenarios {
			Convey(errorScenario.description, func() {
				vaultClient.ReadKeyFunc = validVaultReadFunc

				s3Client.GetWithPSKFunc = validGetWithPSKFunc
				s3Client.FileExistsFunc = fileDoesNotExistFunc
				s3Client.UploadFunc = validUploadFunc

				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(errorScenario.status)
				}))
				defer s.Close()

				err := generateDecrypterCopier(s.URL).HandleFilePublishMessage(ctx, 1, generateMockMessage())

				Convey("Then a No Commit error should be returned", func() {
					ensureNoCommitErrorWithPartMessage(err, errorScenario.errorString)
				})
			})
		}
	})
}

func ensureNoCommitErrorWithPartMessage(err error, part string) {
	So(err, ShouldBeError)
	So(err.Error(), ShouldContainSubstring, part)
	ensureNoCommitError(err)
}

func ensureNoCommitErrorWithMessage(err error, msg string) {
	So(err, ShouldBeError)
	So(err.Error(), ShouldEqual, msg)
	ensureNoCommitError(err)
}

func ensureNoCommitError(err error) {
	commiter, ok := err.(kafka.Commiter)

	So(ok, ShouldBeTrue)
	So(commiter.Commit(), ShouldBeFalse)
}

func generateDecrypterCopier(fileAPIURL string) file.DecrypterCopier {
	return file.NewDecrypterCopier(s3Client, s3Client, vaultClient, vaultPath, fileAPIURL, serviceAuthToken)
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
