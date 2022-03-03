package file

import (
	"context"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
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

	reader, _, _ := d.PrivateClient.GetWithPSK(fp.Path, []byte("1234567890ABCDEF"))
	d.PublicClient.Upload(&s3manager.UploadInput{
		Key:         &fp.Path,
		ContentType: &fp.Type,
		Body:        reader,
	})

	// TODO create a file on public bucket
	return nil
}
