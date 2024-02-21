package event

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-static-file-publisher/schema"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//go:generate moq -out mock/s3_reader.go -pkg mock . S3Reader
//go:generate moq -out mock/s3_writer.go -pkg mock . S3Writer
//go:generate moq -out mock/image_api.go -pkg mock . ImageAPIClient

const (
	failedState    = "failed_publish"
	completedState = "completed"
)

// ImagePublishedHandler hold the details for publishing to s3.
type ImagePublishedHandler struct {
	AuthToken       string
	S3Public        S3Writer
	S3Private       S3Reader
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
}

// ImageAPIClient defines the required methods from image API client
type ImageAPIClient interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	GetImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string) (image.Image, error)
	PutImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string, data image.Image) (image.Image, error)
	GetDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID, variant string) (m image.ImageDownload, err error)
	PutDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID, variant string, data image.ImageDownload) (image.ImageDownload, error)
}

// Handle takes a single event. It writes the file to the public static bucket.
func (h *ImagePublishedHandler) Handle(ctx context.Context, event *ImagePublished) error {
	privateBucket := h.S3Private.BucketName()
	publicBucket := h.S3Public.BucketName()
	logData := log.Data{
		"event":          event,
		"private_bucket": privateBucket,
		"public_bucket":  publicBucket,
	}
	log.Info(ctx, "event handler called", logData)

	// GET images/{id}/downloads/{variant}
	imageDownload, err := h.ImageAPICli.GetDownloadVariant(ctx, "", h.AuthToken, "", event.ImageID, event.ImageVariant)
	if err != nil {
		log.Error(ctx, "error getting image variant from API", err, logData)
		h.setImageStatusToFailed(ctx, event.ImageID, fmt.Sprintf("error getting image variant '%s' from API", event.ImageVariant))
		return err
	}

	privatePath := event.SrcPath

	// Read image from private bucket
	reader, err := h.getS3Reader(privatePath)
	if err != nil {
		log.Error(ctx, "error getting s3 object reader", err, logData)
		h.setVariantStatusToFailed(ctx, event.ImageID, imageDownload, "error getting s3 object reader")
		return err
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
		return err
	}
	endTime := time.Now().UTC()

	imageDownload.State = completedState
	imageDownload.PublishCompleted = &endTime
	imageDownload.Href = fmt.Sprintf("%s/%s", h.PublicBucketURL, event.DstPath)
	imageDownload, err = h.ImageAPICli.PutDownloadVariant(ctx, "", h.AuthToken, "", event.ImageID, event.ImageVariant, imageDownload)
	if err != nil {
		log.Error(ctx, "error putting image variant to API", err, logData)
		h.setImageStatusToFailed(ctx, event.ImageID, fmt.Sprintf("error putting updated image variant '%s' to API", event.ImageVariant))
		return err
	}
	log.Info(ctx, "put image download to api", logData)
	log.Info(ctx, "event successfully handled", logData)
	return nil
}

func (h *ImagePublishedHandler) KafkaHandler(ctx context.Context, msgs []kafka.Message) error {
	logData := log.Data{}
	schem := schema.ImagePublishedEvent
	fp := ImagePublished{}
	log.Info(ctx, fmt.Sprintf("ImagePublishedHandler (batched) invoked with %d message(s)", len(msgs)))

	for _, msg := range msgs {
		if err := schem.Unmarshal(msg.GetData(), &fp); err != nil {
			return fmt.Errorf("ImagePublishedHandler: couldn't unmarshal message: %w", err)
		}

		logData["message"] = fp
		log.Info(ctx, "ImagePublishedHandler: message received", logData)

		err := h.Handle(ctx, &fp)
		if err != nil {
			log.Error(ctx, "ImagePublishedHandler: failed to handle event", err)
			return err
		}
	}
	return nil
}

// Get an S3 reader
func (h *ImagePublishedHandler) getS3Reader(imagePath string) (reader io.ReadCloser, err error) {
	// Get image from upload bucket
	reader, _, err = h.S3Private.Get(imagePath)
	if err != nil {
		return
	}
	return
}

// Upload to public S3 from a reader
func (h *ImagePublishedHandler) uploadToS3(thePath string, reader io.Reader) error {
	publicBucket := h.S3Public.BucketName()
	uploadInput := &s3manager.UploadInput{
		Body:   reader,
		Bucket: &publicBucket,
		Key:    &thePath,
	}

	// Upload file to public bucket
	_, err := h.S3Public.Upload(uploadInput)
	if err != nil {
		return err
	}

	return nil
}

func (h *ImagePublishedHandler) setImageStatusToFailed(ctx context.Context, imageID, desc string) {
	img, err := h.ImageAPICli.GetImage(ctx, "", h.AuthToken, "", imageID)
	if err != nil {
		log.Error(ctx, "error getting image from API to set failed_publish status", err)
		return
	}
	img.State = failedState
	img.Error = desc
	_, err = h.ImageAPICli.PutImage(ctx, "", h.AuthToken, "", imageID, img)
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
