package file

import (
	"context"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	s3client "github.com/ONSdigital/dp-s3/v2"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Published struct {
	Path        string `avro:"path"`
	Type        string `avro:"type"`
	Etag        string `avro:"etag"`
	SizeInBytes string `avro:"sizeInBytes"`
}

type DecrypterCopier struct {
	PublicClient  *s3client.Client
	PrivateClient *s3client.Client
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
