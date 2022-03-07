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
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"strings"
	"testing"
)

type MockMessage struct {
	Data []byte
}

func (m MockMessage) GetData() []byte {
	return m.Data
}

func (m MockMessage) Mark() {
}

func (m MockMessage) Commit() {

}

func (m MockMessage) Release() {
}

func (m MockMessage) CommitAndRelease() {
}

func (m MockMessage) Offset() int64 {
	return 1
}

func (m MockMessage) UpstreamDone() chan struct{} {
	return nil
}

func TestHandleFilePublishMessage(t *testing.T) {
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
	fp := file.Published{
		Path:        "test/file.txt",
		Type:        "plain/test",
		Etag:        "1234567890",
		SizeInBytes: "123",
	}

	c, _ := schema.Marshal(fp)

	msg := MockMessage{
		Data: c,
	}
	ctx := context.Background()

	Convey("Given invalid message content", t, func() {
		dc := file.DecrypterCopier{}
		msg.Data = []byte("Testing")

		Convey("When the message is handled", func() {
			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)

			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})
	})

	Convey("Given there is a read error on the Vault key", t, func() {
		vc := &eventMock.VaultClientMock{ReadKeyFunc: func(path string, key string) (string, error) {
			return "", errors.New("broken")
		}}

		dc := file.DecrypterCopier{
			VaultClient: vc,
		}

		Convey("When the message is handled", func() {
			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})
	})

	Convey("Given there a specific vault read error", t, func() {
		vc := &eventMock.VaultClientMock{}

		dc := file.DecrypterCopier{
			VaultClient: vc,
		}

		msg.Data = c

		Convey("When the error is ErrKeyNotFound", func() {
			vc.ReadKeyFunc = func(path string, key string) (string, error) {
				return "", vault.ErrKeyNotFound
			}

			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, vault.ErrKeyNotFound.Error())
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})

		Convey("When the error is ErrVersionNotFound", func() {
			vc.ReadKeyFunc = func(path string, key string) (string, error) {
				return "", vault.ErrVersionNotFound
			}

			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, vault.ErrVersionNotFound.Error())
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})

		Convey("When the error is ErrMetadataNotFound", func() {
			vc.ReadKeyFunc = func(path string, key string) (string, error) {
				return "", vault.ErrMetadataNotFound
			}

			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, vault.ErrMetadataNotFound.Error())
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})

		Convey("When the error is ErrDataNotFound", func() {
			vc.ReadKeyFunc = func(path string, key string) (string, error) {
				return "", vault.ErrDataNotFound
			}

			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, vault.ErrDataNotFound.Error())
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})

		Convey("When the error is ErrVersionInvalid", func() {
			vc.ReadKeyFunc = func(path string, key string) (string, error) {
				return "", vault.ErrVersionInvalid
			}

			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, vault.ErrVersionInvalid.Error())
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})
	})

	Convey("Given the encryption key from vault cannot be parse to a byte array", t, func() {
		vc := &eventMock.VaultClientMock{}

		dc := file.DecrypterCopier{
			VaultClient: vc,
		}
		msg := MockMessage{
			Data: c,
		}

		Convey("When the encryption key is abcdefgh", func() {
			vc.ReadKeyFunc = func(path string, key string) (string, error) {
				return "abcdefgh", nil
			}

			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, "encoding/hex:")

			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})
	})

	Convey("Given files in the private bucket could not be read", t, func() {
		vc := &eventMock.VaultClientMock{
			ReadKeyFunc: func(path string, key string) (string, error) {
				return "1234567890123456", nil
			},
		}
		const errMsg = "could not read from private bucket"

		pc := &fileMock.S3ClientV2Mock{
			GetWithPSKFunc: func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return nil, nil, errors.New(errMsg)
			},
			UploadFunc: nil,
		}

		dc := file.DecrypterCopier{
			VaultClient:   vc,
			PrivateClient: pc,
		}

		Convey("Attempting to get a file from to  private s3 bucket", func() {
			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, errMsg)
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})
	})

	Convey("Given files sent the public bucket could not be written", t, func() {
		vc := &eventMock.VaultClientMock{
			ReadKeyFunc: func(path string, key string) (string, error) {
				return "1234567890123456", nil
			},
		}
		const errMsg = "could not write to public bucket"

		pc := &fileMock.S3ClientV2Mock{
			GetWithPSKFunc: func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return io.NopCloser(strings.NewReader("testing")), nil, nil
			},
			UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return nil, errors.New(errMsg)
			},
		}

		dc := file.DecrypterCopier{
			VaultClient:   vc,
			PrivateClient: pc,
			PublicClient:  pc,
		}

		Convey("Attempting to get a file from to  private s3 bucket", func() {
			err := dc.HandleFilePublishMessage(ctx, 1, msg)

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, errMsg)
			commiter, ok := err.(kafka.Commiter)

			So(ok, ShouldBeTrue)
			So(commiter.Commit(), ShouldBeFalse)
		})
	})

}
