package event_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"github.com/ONSdigital/dp-api-clients-go/image"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/dp-static-file-publisher/event/mock"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testVaultPath            = "/vault/path/for/testing"
	testVaultPrivateFilePath = "/vault/path/for/testing/images/ID1/original"
	testAuthToken            = "auth-123"
	testPublicBucketURL      = "http://some.bucket.url"
	failedState              = "failed_publish"
)

var (
	testPrivateBucket string        = "privateBucket"
	testPublicBucket  string        = "publicBucket"
	testSize          int64         = 1234
	fileBytes         []byte        = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	testFileContent   io.ReadCloser = ioutil.NopCloser(bytes.NewReader(fileBytes))
	encodedPSK        string        = "48656C6C6F20576F726C64"
	errVault                        = errors.New("Vault error")
	errS3Private                    = errors.New("S3Private error")
	errS3Public                     = errors.New("S3Public error")
	errImageAPI                     = errors.New("imageAPI error")

	testPublishStarted   = time.Date(2020, time.April, 26, 8, 5, 52, 0, time.UTC)
	testPublishCompleted = time.Date(2020, time.April, 26, 8, 7, 32, 0, time.UTC)

	testPublishedDownload = image.ImageDownload{
		Id:             "original",
		State:          "published",
		PublishStarted: &testPublishStarted,
	}

	testCompletedDownload = image.ImageDownload{
		Id:               "original",
		State:            "completed",
		PublishStarted:   &testPublishStarted,
		PublishCompleted: &testPublishCompleted,
	}
)

var testEventNoSrcPath = event.ImagePublished{}

func TestImagePublishedHandler_Handle(t *testing.T) {

	Convey("Given S3 and Vault client mocks", t, func() {

		mockS3Private := &mock.S3ReaderMock{
			BucketNameFunc: func() string {
				return testPrivateBucket
			},
		}
		mockS3Public := &mock.S3WriterMock{
			BucketNameFunc: func() string {
				return testPublicBucket
			},
		}
		mockVault := &mock.VaultClientMock{
			ReadKeyFunc: func(path string, key string) (string, error) {
				return encodedPSK, nil
			},
		}
		mockImageAPI := &mock.ImageAPIClientMock{
			GetImageFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string) (image.Image, error) {
				return image.Image{}, nil
			},
			PutImageFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.Image) (image.Image, error) {
				return data, nil
			},
			GetDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string) (image.ImageDownload, error) {
				return testPublishedDownload, nil
			},
			PutDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string, data image.ImageDownload) (image.ImageDownload, error) {
				return testCompletedDownload, nil
			},
		}
		psk, err := hex.DecodeString(encodedPSK)
		So(err, ShouldBeNil)

		Convey("And a successful event handler, when Handle is triggered", func() {
			mockS3Private.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Public.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken:       testAuthToken,
				S3Public:        mockS3Public,
				S3Private:       mockS3Private,
				VaultCli:        mockVault,
				VaultPath:       testVaultPath,
				ImageAPICli:     mockImageAPI,
				PublicBucketURL: testPublicBucketURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)
			So(err, ShouldBeNil)

			Convey("Encryption key is read from Vault with the expected path", func() {
				So(mockVault.ReadKeyCalls(), ShouldHaveLength, 1)
				So(mockVault.ReadKeyCalls()[0].Path, ShouldEqual, testVaultPrivateFilePath)
				So(mockVault.ReadKeyCalls()[0].Key, ShouldEqual, "key")
			})

			Convey("The file is obtained from the private bucket and decrypted with the psk obtained from Vault", func() {
				So(mockS3Private.GetWithPSKCalls(), ShouldHaveLength, 1)
				So(mockS3Private.GetWithPSKCalls()[0].Key, ShouldEqual, testEvent.SrcPath)
				So(mockS3Private.GetWithPSKCalls()[0].Psk, ShouldResemble, psk)
			})

			Convey("An image download variant is retrieved from the image API", func() {
				So(mockImageAPI.GetDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.GetDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
			})

			Convey("The file is uploaded to the public bucket", func() {
				So(mockS3Public.UploadCalls(), ShouldHaveLength, 1)
				So(*mockS3Public.UploadCalls()[0].Input, ShouldResemble, s3manager.UploadInput{
					Body:   testFileContent,
					Bucket: &testPublicBucket,
					Key:    &testEvent.DstPath,
				})
			})

			Convey("The image download variant is put to the image API with a state of completed", func() {
				So(mockImageAPI.PutDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
				newImageData := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(newImageData, ShouldNotBeNil)
				So(newImageData.Id, ShouldEqual, "original")
				So(newImageData.State, ShouldEqual, "completed")
				So(newImageData.PublishCompleted, ShouldNotBeNil)
				So(newImageData.Href, ShouldEqual, testPublicBucketURL+"/"+testEvent.DstPath)
			})
		})

		Convey("And a nil-vault event handler (developer env), when Handle is triggered", func() {
			mockS3Private.GetFunc = func(key string) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Public.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken:       testAuthToken,
				S3Public:        mockS3Public,
				S3Private:       mockS3Private,
				VaultCli:        nil,
				VaultPath:       "",
				ImageAPICli:     mockImageAPI,
				PublicBucketURL: testPublicBucketURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)
			So(err, ShouldBeNil)

			Convey("The file is obtained from the private bucket and decrypted with the psk obtained from Vault", func() {
				So(mockS3Private.GetCalls(), ShouldHaveLength, 1)
				So(mockS3Private.GetCalls()[0].Key, ShouldEqual, testEvent.SrcPath)
			})

			Convey("An image download variant is retrieved from the image API", func() {
				So(mockImageAPI.GetDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.GetDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
			})

			Convey("The file is uploaded to the public bucket", func() {
				So(mockS3Public.UploadCalls(), ShouldHaveLength, 1)
				So(*mockS3Public.UploadCalls()[0].Input, ShouldResemble, s3manager.UploadInput{
					Body:   testFileContent,
					Bucket: &testPublicBucket,
					Key:    &testEvent.DstPath,
				})
			})

			Convey("The image download variant is put to the image API with a state of completed", func() {
				So(mockImageAPI.PutDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ImageID, ShouldEqual, testEvent.ImageID)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)
				newImageData := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(newImageData, ShouldNotBeNil)
				So(newImageData.Id, ShouldEqual, "original")
				So(newImageData.State, ShouldEqual, "completed")
				So(newImageData.PublishCompleted, ShouldNotBeNil)
				So(newImageData.Href, ShouldEqual, testPublicBucketURL+"/"+testEvent.DstPath)
			})
		})

		Convey("And an event handler with a failing vault client, when Handle is triggered", func() {
			mockVaultFail := &mock.VaultClientMock{
				ReadKeyFunc: func(path string, key string) (string, error) {
					return "", errVault
				},
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken:   testAuthToken,
				S3Public:    mockS3Public,
				S3Private:   mockS3Private,
				VaultCli:    mockVaultFail,
				VaultPath:   testVaultPath,
				ImageAPICli: mockImageAPI,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("Vault ReadKey is called and the error is returned", func() {
				So(err, ShouldResemble, errVault)
				So(mockVaultFail.ReadKeyCalls(), ShouldHaveLength, 1)
			})

			Convey("The download variant is retrieved from the API and updated with a state of failed_publish", func() {
				So(mockImageAPI.GetDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.GetDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				So(mockImageAPI.PutDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				updatedImage := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(updatedImage.State, ShouldEqual, failedState)
				So(updatedImage.Error, ShouldEqual, "error reading key from vault")
			})
		})

		Convey("And an event handler with a vault client that returns an invalid psk, when Handle is triggered", func() {
			mockVaultFail := &mock.VaultClientMock{
				ReadKeyFunc: func(path string, key string) (string, error) {
					return "invalidPSK", nil
				},
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken:   testAuthToken,
				S3Public:    mockS3Public,
				S3Private:   mockS3Private,
				VaultCli:    mockVaultFail,
				VaultPath:   testVaultPath,
				ImageAPICli: mockImageAPI,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("Vault ReadKey is called and the decoding error is returned", func() {
				So(err, ShouldNotBeNil)
				So(mockVaultFail.ReadKeyCalls(), ShouldHaveLength, 1)
			})

			Convey("The download variant is retrieved from the API and updated with a state of failed_publish", func() {
				So(mockImageAPI.GetDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.GetDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				So(mockImageAPI.PutDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				updatedImage := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(updatedImage.State, ShouldEqual, failedState)
				So(updatedImage.Error, ShouldEqual, "error reading key from vault")
			})
		})

		Convey("And an event handler with an S3Private client that fails to obtain the source file, when Handle is triggered", func() {
			mockS3Private.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return nil, nil, errS3Private
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken:   testAuthToken,
				S3Public:    mockS3Public,
				S3Private:   mockS3Private,
				VaultCli:    mockVault,
				VaultPath:   testVaultPath,
				ImageAPICli: mockImageAPI,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("S3Private is called and the same error is returned", func() {
				So(err, ShouldResemble, errS3Private)
				So(mockVault.ReadKeyCalls(), ShouldHaveLength, 1)
				So(mockS3Private.GetWithPSKCalls(), ShouldHaveLength, 1)
			})

			Convey("The download variant is retrieved from the API and updated with a state of failed_publish", func() {
				So(mockImageAPI.GetDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.GetDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				So(mockImageAPI.PutDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				updatedImage := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(updatedImage.State, ShouldEqual, failedState)
				So(updatedImage.Error, ShouldEqual, "error getting s3 object reader")
			})
		})

		Convey("And a nil-vault event handler (developer env) with an S3Uploaded client that fails to obtain the source file, when Handle is triggered", func() {
			mockS3Private.GetFunc = func(key string) (io.ReadCloser, *int64, error) {
				return nil, nil, errS3Private
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken: 	 testAuthToken,
				S3Public:    mockS3Public,
				S3Private:   mockS3Private,
				VaultCli:    nil,
				VaultPath:   "",
				ImageAPICli: mockImageAPI,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("S3Private is called and the same error is returned", func() {
				So(err, ShouldResemble, errS3Private)
				So(mockS3Private.GetCalls(), ShouldHaveLength, 1)
			})

			Convey("The download variant is retrieved from the API and updated with a state of failed_publish", func() {
				So(mockImageAPI.GetDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.GetDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.GetDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				So(mockImageAPI.PutDownloadVariantCalls(), ShouldHaveLength, 1)
				So(mockImageAPI.PutDownloadVariantCalls()[0].Variant, ShouldEqual, testEvent.ImageVariant)
				So(mockImageAPI.PutDownloadVariantCalls()[0].ServiceAuthToken, ShouldResemble, testAuthToken)

				updatedImage := mockImageAPI.PutDownloadVariantCalls()[0].Data
				So(updatedImage.State, ShouldEqual, failedState)
				So(updatedImage.Error, ShouldEqual, "error getting s3 object reader")
			})
		})

		Convey("And an event handler with an image client that fails to retrieve a variant, when Handle is triggered", func() {
			mockS3Private.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockImageAPIFail := &mock.ImageAPIClientMock{
				GetDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string) (image.ImageDownload, error) {
					return image.ImageDownload{}, errImageAPI
				},
				GetImageFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string) (image.Image, error) {
					return image.Image{}, errImageAPI
				},
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken:       testAuthToken,
				S3Public:        mockS3Public,
				S3Private:       mockS3Private,
				VaultCli:        mockVault,
				VaultPath:       testVaultPath,
				ImageAPICli:     mockImageAPIFail,
				PublicBucketURL: testPublicBucketURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("ImageAPI.PostDownloadVariant is called and the error is returned", func() {
				So(err, ShouldNotBeNil)
				So(mockImageAPIFail.GetDownloadVariantCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("And an event handler with an S3Public client that fails to upload the file, when Handle is triggered", func() {
			mockS3Private.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Public.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return nil, errS3Public
			}
			eventHandler := event.ImagePublishedHandler{
				S3Public:    mockS3Public,
				S3Private:   mockS3Private,
				VaultCli:    mockVault,
				VaultPath:   testVaultPath,
				ImageAPICli: mockImageAPI,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("S3Private is called and the same error is returned", func() {
				So(err, ShouldResemble, errS3Public)
				So(mockVault.ReadKeyCalls(), ShouldHaveLength, 1)
				So(mockS3Private.GetWithPSKCalls(), ShouldHaveLength, 1)
				So(mockS3Public.BucketNameCalls(), ShouldHaveLength, 2)
			})
		})

		Convey("And an event handler with an image client that fails to update a variant, when Handle is triggered", func() {
			mockS3Private.GetWithPSKFunc = func(key string, psk []byte) (io.ReadCloser, *int64, error) {
				return testFileContent, &testSize, nil
			}
			mockS3Public.UploadFunc = func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				return &s3manager.UploadOutput{}, nil
			}
			mockImageAPIFail := &mock.ImageAPIClientMock{
				GetImageFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string) (image.Image, error) {
					return image.Image{}, errImageAPI
				},
				GetDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string) (image.ImageDownload, error) {
					return testPublishedDownload, nil
				},
				PutDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string, data image.ImageDownload) (image.ImageDownload, error) {
					return image.ImageDownload{}, errImageAPI
				},
			}
			eventHandler := event.ImagePublishedHandler{
				AuthToken:       testAuthToken,
				S3Public:        mockS3Public,
				S3Private:       mockS3Private,
				VaultCli:        mockVault,
				VaultPath:       testVaultPath,
				ImageAPICli:     mockImageAPIFail,
				PublicBucketURL: testPublicBucketURL,
			}
			err := eventHandler.Handle(testCtx, &testEvent)

			Convey("ImageAPI.PutDownloadVariant is called and the error is returned", func() {
				So(err, ShouldNotBeNil)
				So(mockImageAPIFail.PutDownloadVariantCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("And an event with no path supplied, when Handle is triggered", func() {
			eventHandler := event.ImagePublishedHandler{
				AuthToken:       testAuthToken,
				S3Private:       mockS3Private,
				S3Public:        mockS3Public,
				VaultCli:        mockVault,
				VaultPath:       testVaultPath,
				ImageAPICli:     mockImageAPI,
				PublicBucketURL: testPublicBucketURL,
			}
			err := eventHandler.Handle(testCtx, &testEventNoSrcPath)

			Convey("The image download variant is put to the image API with a state of imported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "vault filename required but was empty")
			})
		})

	})

}
