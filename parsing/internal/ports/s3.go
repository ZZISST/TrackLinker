package ports

type S3Port interface {
	UploadTrackPreview(mp3URL, title, artistNames string) (previewAudioUrl *string, err error)
}
