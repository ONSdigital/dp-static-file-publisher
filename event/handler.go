package event

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//go:generate moq -out mock/s3_reader.go -pkg mock . S3Reader
//go:generate moq -out mock/s3_writer.go -pkg mock . S3Writer
//go:generate moq -out mock/vault.go -pkg mock . VaultClient
//go:generate moq -out mock/image_api.go -pkg mock . ImageAPIClient

const (
	vaultKey       = "key" // is the key under each vault secret that contains the PSK needed to decrypt files from S3
	failedState    = "failed_publish"
	completedState = "completed"
)

// ErrVaultFilenameEmpty is an error returned when trying to obtain a PSK for an empty file name
var ErrVaultFilenameEmpty = errors.New("vault filename required but was empty")

// ImagePublishedHandler ...
type ImagePublishedHandler struct {
	AuthToken       string
	S3Public        S3Writer
	S3Private       S3Reader
	VaultCli        VaultClient
	VaultPath       string
	ImageAPICli     ImageAPIClient
	PublicBucketURL string
}

// S3Writer defines the required methods from dp-s3 to interact with a particular bucket of AWS S3
type S3Writer interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// S3Reader defines the required methods from dp-s3 to read data to an AWS S3 Bucket
type S3Reader interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	Get(key string) (io.ReadCloser, *int64, error)
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
}

// VaultClient defines the required methods from dp-vault client
type VaultClient interface {
	ReadKey(path, key string) (string, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// ImageAPIClient defines the required methods from image API client
type ImageAPIClient interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	GetImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string) (image.Image, error)
	PutImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string, data image.Image) (image.Image, error)
	GetDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID, variant string) (m image.ImageDownload, err error)
	PutDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID, variant string, data image.ImageDownload) (image.ImageDownload, error)
}

// Handle takes a single event. It reads the PSK from Vault, uses it to decrypt the encrypted file
// from the private S3 bucket, and writes it to the public static bucket without using the vault psk for encryption.
func (h *ImagePublishedHandler) Handle(ctx context.Context, event *ImagePublished) (err error) {
	privateBucket := h.S3Private.BucketName()
	publicBucket := h.S3Public.BucketName()
	logData := log.Data{
		"event":          event,
		"private_bucket": privateBucket,
		"public_bucket":  publicBucket,
		"vault_path":     h.VaultPath,
	}
	log.Info(ctx, "event handler called", logData)

	// GET images/{id}/downloads/{variant}
	imageDownload, err := h.ImageAPICli.GetDownloadVariant(ctx, "", h.AuthToken, "", event.ImageID, event.ImageVariant)
	if err != nil {
		log.Error(ctx, "error getting image variant from API", err, logData)
		h.setImageStatusToFailed(ctx, event.ImageID, fmt.Sprintf("error getting image variant '%s' from API", event.ImageVariant))
		return
	}

	privatePath := event.SrcPath

	// Get PSK from Vault
	var privatePsk []byte
	if h.VaultCli != nil {
		privatePsk, err = h.getVaultKeyForFile(privatePath)
		if err != nil {
			log.Error(ctx, "error reading key from vault", err, logData)
			h.setVariantStatusToFailed(ctx, event.ImageID, imageDownload, "error reading key from vault")
			return err
		}
	}

	// Decrypt image from private bucket optionally using PSK obtained from Vault
	reader, err := h.getS3Reader(privatePath, privatePsk)
	if err != nil {
		log.Error(ctx, "error getting s3 object reader", err, logData)
		h.setVariantStatusToFailed(ctx, event.ImageID, imageDownload, "error getting s3 object reader")
		return
	}
	defer reader.Close()

	logData["imageDownload"] = &imageDownload
	log.Info(ctx, "got image download from api", logData)

	// Upload file to public bucket
	log.Info(ctx, "uploading private file to s3", logData)
	err = h.uploadToS3(event.DstPath, reader)
	if err != nil {
		log.Error(ctx, "error uploading to s3", err, logData)
		h.setVariantStatusToFailed(ctx, event.ImageID, imageDownload, "failed to upload image to s3")
		return
	}
	endTime := time.Now().UTC()

	imageDownload.State = completedState
	imageDownload.PublishCompleted = &endTime
	imageDownload.Href = fmt.Sprintf("%s/%s", h.PublicBucketURL, event.DstPath)
	imageDownload, err = h.ImageAPICli.PutDownloadVariant(ctx, "", h.AuthToken, "", event.ImageID, event.ImageVariant, imageDownload)
	if err != nil {
		log.Error(ctx, "error putting image variant to API", err, logData)
		h.setImageStatusToFailed(ctx, event.ImageID, fmt.Sprintf("error putting updated image variant '%s' to API", event.ImageVariant))
		return
	}
	log.Info(ctx, "put image download to api", logData)
	log.Info(ctx, "event successfully handled", logData)
	return nil
}

// Get an S3 reader
func (h *ImagePublishedHandler) getS3Reader(path string, psk []byte) (reader io.ReadCloser, err error) {
	if psk != nil {
		// Decrypt image from upload bucket using PSK obtained from Vault
		reader, _, err = h.S3Private.GetWithPSK(path, psk)
		if err != nil {
			return
		}
	} else {
		// Get image from upload bucket
		reader, _, err = h.S3Private.Get(path)
		if err != nil {
			return
		}
	}
	return
}

// Upload to public S3 from a reader
func (h *ImagePublishedHandler) uploadToS3(path string, reader io.Reader) error {
	publicBucket := h.S3Public.BucketName()
	uploadInput := &s3manager.UploadInput{
		Body:   reader,
		Bucket: &publicBucket,
		Key:    &path,
	}

	// Upload file to public bucket
	_, err := h.S3Public.Upload(uploadInput)
	if err != nil {
		return err
	}

	return nil
}

// getVaultKeyForFile reads the encryption key from Vault for the provided path
func (h *ImagePublishedHandler) getVaultKeyForFile(keyPath string) ([]byte, error) {
	if len(keyPath) == 0 {
		return nil, ErrVaultFilenameEmpty
	}

	vp := path.Join(h.VaultPath, keyPath)
	pskStr, err := h.VaultCli.ReadKey(vp, vaultKey)
	if err != nil {
		return nil, err
	}

	psk, err := hex.DecodeString(pskStr)
	if err != nil {
		return nil, err
	}

	return psk, nil
}

func (h *ImagePublishedHandler) setImageStatusToFailed(ctx context.Context, imageID string, desc string) {
	image, err := h.ImageAPICli.GetImage(ctx, "", h.AuthToken, "", imageID)
	if err != nil {
		log.Error(ctx, "error getting image from API to set failed_publish status", err)
		return
	}
	image.State = failedState
	image.Error = desc
	_, err = h.ImageAPICli.PutImage(ctx, "", h.AuthToken, "", imageID, image)
	if err != nil {
		log.Error(ctx, "error putting image to API to set failed_publish  status", err)
		return
	}
}

func (h *ImagePublishedHandler) setVariantStatusToFailed(ctx context.Context, imageID string, variant image.ImageDownload, desc string) {
	variant.State = failedState
	variant.Error = desc
	_, err := h.ImageAPICli.PutDownloadVariant(ctx, "", h.AuthToken, "", imageID, variant.Id, variant)
	if err != nil {
		log.Error(ctx, "error putting image variant to API to set failed_publish status", err)
		return
	}
}
