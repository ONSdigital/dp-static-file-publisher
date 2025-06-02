package schema

import (
	"github.com/ONSdigital/dp-kafka/v4/avro"
)

var imagePublishedEvent = `{
  "type": "record",
  "name": "image-published",
  "fields": [
    {"name": "src_path", "type": "string", "default": ""},
    {"name": "dst_path", "type": "string", "default": ""},
    {"name": "image_id", "type": "string", "default": ""},
    {"name": "image_variant", "type": "string", "default": ""}
  ]
}`

// ImagePublishedEvent is the Avro schema for Image published messages.
var ImagePublishedEvent = &avro.Schema{
	Definition: imagePublishedEvent,
}
