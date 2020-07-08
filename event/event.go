package event

// ImagePublished provides an avro structure for an image published event
type ImagePublished struct {
	ImageID string `avro:"image_id"`
}
