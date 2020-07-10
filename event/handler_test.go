package event_test

import (
	"bytes"
	"encoding/hex"
	"io"
	"io/ioutil"
	"testing"

	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testVaultPath = "/vault/path/for/testing"
)

var (
	testPrivateBucket string        = "privateBucket"
	testPublicBucket  string        = "publicBucket"
	testSize          int64         = 1234
	fileBytes         []byte        = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	testFileContent   io.ReadCloser = ioutil.NopCloser(bytes.NewReader(fileBytes))
	encodedPSK        string        = "48656C6C6F20576F726C64"
)

func TestDataImportCompleteHandler_Handle_HierarchyStoreError(t *testing.T) {

	Convey("Given an event handler with successful mock S3 and Vault clients", t, func() {

		mockS3Private := &mock.S3ClientMock{
			BucketNameFunc: func() string {
				return testPrivateBucket
			},
			GetWithPSKFunc: func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			},
		}

		mockS3Public := &mock.S3UploaderMock{
			BucketNameFunc: func() string {
				return testPublicBucket
			},
			UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			},
		}

		mockVault := &mock.VaultClientMock{
			ReadKeyFunc: func(path string, key string) (string, error) {
				return encodedPSK, nil
			},
		}

		eventHandler := event.ImagePublishedHandler{
			S3Public:  mockS3Public,
			S3Private: mockS3Private,
			VaultCli:  mockVault,
			VaultPath: testVaultPath,
		}

		psk, err := hex.DecodeString(encodedPSK)
		So(err, ShouldBeNil)

		Convey("When Handle is triggered", func() {
			err := eventHandler.Handle(testCtx, &testEvent)
			So(err, ShouldBeNil)

			Convey("Encryption key is read from Vault with the expected path", func() {
				So(len(mockVault.ReadKeyCalls()), ShouldEqual, 1)
				So(mockVault.ReadKeyCalls()[0].Path, ShouldEqual, testVaultPath)
				So(mockVault.ReadKeyCalls()[0].Key, ShouldEqual, "key")
			})

			Convey("The file is obtained from the private bucket and decrypted with the psk obtained from Vault", func() {
				So(len(mockS3Private.GetWithPSKCalls()), ShouldEqual, 1)
				So(mockS3Private.GetWithPSKCalls()[0].Key, ShouldEqual, testEvent.SrcPath)
				So(mockS3Private.GetWithPSKCalls()[0].Psk, ShouldResemble, psk)
			})

			Convey("The file is uploaded to the public bucket", func() {
				So(len(mockS3Public.UploadCalls()), ShouldEqual, 1)
				So(*mockS3Public.UploadCalls()[0].Input, ShouldResemble, s3manager.UploadInput{
					Body:   testFileContent,
					Bucket: &testPublicBucket,
					Key:    &testEvent.DstPath,
				})
			})
		})
	})

	// TODO implement negative path tests
}
