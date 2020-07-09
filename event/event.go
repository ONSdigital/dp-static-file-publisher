package event

// ImagePublished provides an avro structure for an image published event
type ImagePublished struct {
	Downloads []ImageDownloadPublished `avro:"downloads"`
}

// ImageDownloadPublished provides an avro structure for each download of an image published event
type ImageDownloadPublished struct {
	SrcPath string `avro:"src_path"`
	DstPath string `avro:"dst_path"`
}
