package file

import (
	"context"
	"encoding/hex"
	"fmt"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-static-file-publisher/event"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type Published struct {
	Path        string `avro:"path"`
	Type        string `avro:"type"`
	Etag        string `avro:"etag"`
	SizeInBytes string `avro:"sizeInBytes"`
}

//go:generate moq -out mock/s3client.go -pkg mock . S3ClientV2
type S3ClientV2 interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	GetWithPSK(key string, psk []byte) (io.ReadCloser, *int64, error)
}

type DecrypterCopier struct {
	PublicClient  S3ClientV2
	PrivateClient S3ClientV2
	VaultClient   event.VaultClient
	VaultPath     string
}

func (d DecrypterCopier) HandleFilePublishMessage(ctx context.Context, workerID int, msg kafka.Message) error {
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
	fp := Published{}
	schema.Unmarshal(msg.GetData(), &fp)

	encryptionKey, err := d.VaultClient.ReadKey(fmt.Sprintf("%s/%s", d.VaultPath, fp.Path), "key")
	if err != nil {
		log.Error(ctx, "getting encryption key", err)
	}
	decodeString, err := hex.DecodeString(encryptionKey)
	if err != nil {
		log.Error(ctx, "decoding encryption key", err)
	}

	reader, _, err := d.PrivateClient.GetWithPSK(fp.Path, decodeString)
	if err != nil {
		log.Error(ctx, "READ ERROR", err)
	}
	_, err = d.PublicClient.Upload(&s3manager.UploadInput{
		Key:         &fp.Path,
		ContentType: &fp.Type,
		Body:        reader,
	})
	if err != nil {
		log.Error(ctx, "WRITE ERROR", err)
	}

	// TODO create a file on public bucket
	return nil
}
