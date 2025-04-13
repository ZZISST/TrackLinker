package service

import "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"

type ICloudService interface {
	UploadTrackPreview(mp3URL, title, artistNames string) (previewAudioUrl *string, err error)
}

type YandexCloudService struct {
	s3 ports.S3Port
}

func (s *YandexCloudService) UploadTrackPreview(mp3URL, title, artistNames string) (previewAudioUrl *string, err error) {
	previewAudioUrl, err = s.s3.UploadTrackPreview(mp3URL, title, artistNames)
	if err != nil {
		return nil, err
	}
	return previewAudioUrl, nil
}

func NewYandexCloudService(s3 ports.S3Port) ICloudService {
	return &YandexCloudService{
		s3: s3,
	}
}
