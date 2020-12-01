package event

// ImagePublished provides an avro structure for an image published event
type ImagePublished struct {
	SrcPath      string `avro:"src_path"`
	DstPath      string `avro:"dst_path"`
	ImageID      string `avro:"image_id"`
	ImageVariant string `avro:"image_variant"`
}
