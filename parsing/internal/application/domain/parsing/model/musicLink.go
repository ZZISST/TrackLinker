package model

import "github.com/google/uuid"

type MusicLink struct {
	ID      string       `json:"id" gorm:"unique;not null;index;primary_key"`
	Service MusicService `json:"music_service"`
	URL     string       `json:"url"`

	// One music link belongs to one track
	TrackID string `json:"track_id" gorm:"index"`
	Track   Track  `json:"track" gorm:"constraint:OnDelete:CASCADE;"`
}

func NewMusicLink(service MusicService, url string, trackId string) MusicLink {
	return MusicLink{
		ID:      uuid.New().String(),
		Service: service,
		URL:     url,
		TrackID: trackId,
	}
}
