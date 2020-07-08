package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var imagePublishedEvent = `{
  "type": "record",
  "name": "image-published",
  "fields": [
    {"name": "image_id", "type": "string", "default": ""}
  ]
}`

// ImagePublishedEvent is the Avro schema for Image published messages.
var ImagePublishedEvent = &avro.Schema{
	Definition: imagePublishedEvent,
}