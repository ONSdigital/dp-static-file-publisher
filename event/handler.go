package event

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
	"github.com/aws/aws-sdk-go/aws/session"
)

//go:generate moq -out mock/s3.go -pkg mock . S3Client

// ImagePublishedHandler ...
type ImagePublishedHandler struct {
	S3Public  S3Client
	S3Private S3Client
}

// S3Client defines the required methods from dp-s3 to interact with a particular bucket of AWS S3
type S3Client interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	Session() *session.Session
}

// Handle takes a single event, and returns the observations gathered from the URL in the event.
func (h ImagePublishedHandler) Handle(ctx context.Context, event *ImagePublished) error {
	logData := log.Data{"event": event}
	log.Event(ctx, "event handler called", log.INFO, logData)

	// TODO decrypt image from private bucket and copy it to static bucket

	log.Event(ctx, "event successfully handled", log.INFO, logData)
	return nil
}
