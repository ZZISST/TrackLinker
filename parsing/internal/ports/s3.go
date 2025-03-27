package ports

type S3Port interface {
	UploadTrackPreview(mp3URL, title, artist string) (previewAudioUrl *string, err error)
}
