package event

import (
	"context"
	"encoding/hex"
	"io"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//go:generate moq -out mock/s3_client.go -pkg mock . S3Client
//go:generate moq -out mock/s3_uploader.go -pkg mock . S3Uploader
//go:generate moq -out mock/vault.go -pkg mock . VaultClient

// VaultKey is the key under the vault that contains the PSK needed to decrypt files from the encrypted private S3 bucket
const VaultKey = "key"

// ImagePublishedHandler ...
type ImagePublishedHandler struct {
	S3Public  S3Uploader
	S3Private S3Client
	VaultCli  VaultClient
	VaultPath string
}

// S3Client defines the required methods from dp-s3 to interact with a particular bucket of AWS S3
type S3Client interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
}

// S3Uploader defines the required methods from dp-s3 to upload data to an AWS S3 Bucket
type S3Uploader interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
	BucketName() string
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// VaultClient defines the required methods from dp-vault client
type VaultClient interface {
	ReadKey(path, key string) (string, error)
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// Handle takes a single event. It reads the PSK from Vault, uses it to decrypt the encrypted file
// from the private S3 bucket, and writes it to the public static bucket without using the vault psk for encryption.
func (h ImagePublishedHandler) Handle(ctx context.Context, event *ImagePublished) error {
	publicBucket := h.S3Public.BucketName()
	logData := log.Data{
		"event":          event,
		"private_bucket": h.S3Private.BucketName(),
		"public_bucket":  publicBucket,
		"vault_path":     h.VaultPath,
	}
	log.Event(ctx, "event handler called", log.INFO, logData)

	// Get PSK from Vault
	pskStr, err := h.VaultCli.ReadKey(h.VaultPath, VaultKey)
	if err != nil {
		log.Event(ctx, "error reading key from vault", log.ERROR, log.Error(err), logData)
		return err
	}
	psk, err := hex.DecodeString(pskStr)
	if err != nil {
		log.Event(ctx, "error decoding vault key", log.ERROR, log.Error(err), logData)
		return err
	}

	for _, eventVariant := range event.Downloads {

		// Decrypt image from private bucket using PSK obtained from Vault
		reader, _, err := h.S3Private.GetWithPSK(eventVariant.SrcPath, psk)
		if err != nil {
			logData["failed_source_path"] = eventVariant.SrcPath
			log.Event(ctx, "error getting s3 object with psk", log.ERROR, log.Error(err), logData)
			return err
		}

		// Upload file to public bucket
		log.Event(ctx, "uploading public file to s3", log.INFO, logData)
		_, err = h.S3Public.Upload(&s3manager.UploadInput{
			Body:   reader,
			Bucket: &publicBucket,
			Key:    &eventVariant.DstPath,
		})
		if err != nil {
			logData["failed_destination_path"] = eventVariant.DstPath
			log.Event(ctx, "error uploading s3 object", log.ERROR, log.Error(err), logData)
			return err
		}
	}

	log.Event(ctx, "event successfully handled", log.INFO, logData)
	return nil
}
