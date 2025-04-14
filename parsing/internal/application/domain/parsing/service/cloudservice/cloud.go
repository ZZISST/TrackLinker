/*
Package service provides implementations for cloud service interactions.

Authors:

	Vladimir Stabrovskiy @godnesty
*/
package cloudservice

import "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"

/*
ICloudService defines the interface for cloud service interactions, such as uploading track previews.

Author: Vladimir Stabrovskiy @godnesty
*/
type ICloudService interface {
	UploadTrackPreview(mp3URL, title, artistNames string) (previewAudioUrl *string, err error)
}

/*
YandexCloudService implements the ICloudService interface using Yandex Cloud's S3 storage.

Author: Vladimir Stabrovskiy @godnesty
*/
type YandexCloudService struct {
	s3 ports.S3Port
}

/*
UploadTrackPreview uploads a track preview to Yandex Cloud's S3 storage.

Parameters:
  - mp3URL: The URL of the MP3 file to upload.
  - title: The title of the track.
  - artistNames: A comma-separated string of artist names.

Returns:
  - previewAudioUrl: The URL of the uploaded preview audio file.
  - err: An error if the upload fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *YandexCloudService) UploadTrackPreview(mp3URL, title, artistNames string) (previewAudioUrl *string, err error) {
	previewAudioUrl, err = s.s3.UploadTrackPreview(mp3URL, title, artistNames)
	if err != nil {
		return nil, err
	}
	return previewAudioUrl, nil
}

/*
NewYandexCloudService creates and returns a new instance of YandexCloudService.

Parameters:
  - s3: An implementation of the S3Port interface for interacting with S3 storage.

Returns:
  - ICloudService: A new instance of YandexCloudService.

Author: Vladimir Stabrovskiy @godnesty
*/
func NewYandexCloudService(s3 ports.S3Port) ICloudService {
	return &YandexCloudService{
		s3: s3,
	}
}
