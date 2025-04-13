package model

import (
	"time"

	"github.com/google/uuid"
)

type Track struct {
	ID            string       `json:"id" gorm:"unique;not null;index;primary_key"`
	Title         string       `json:"title"`
	DurationSec   int          `json:"duration_sec"`
	TrackPosition int          `json:"track_number"`
	DiscNumber    int          `json:"disc_number"`
	ReleaseDate   time.Time    `json:"release_date"`
	TrackPreview  TrackPreview `json:"track_preview" gorm:"foreignKey:TrackID"`
	Artists       []Artist     `json:"artists" gorm:"many2many:track_artists;"`
	AlbumID       string       `json:"album_id" gorm:"index"` // Foreign key
	Album         Album        `json:"album" gorm:"constraint:OnDelete:CASCADE;"`
	Type          string       `json:"type"`
	MusicLinks    []MusicLink  `json:"music_links" gorm:"polymorphicType:OwnerType;polymorphicId:OwnerID;polymorphicValue:track"`
}

type TrackPreview struct {
	ID      string `json:"id" gorm:"unique;not null;index;primary_key"`
	FileURL string `json:"file_url"`

	TrackID string `json:"track_id" gorm:"index:idx_tracks_track_preview"`
}

func NewTrack(title string, durationSec int, trackPosition int, discNumber int, releaseDate time.Time) Track {
	return Track{
		ID:            uuid.New().String(),
		Title:         title,
		DurationSec:   durationSec,
		TrackPosition: trackPosition,
		DiscNumber:    discNumber,
		ReleaseDate:   releaseDate,
		Type:          "track",
	}
}

func NewTrackPreview(trackId string, fileURL string) TrackPreview {
	return TrackPreview{
		ID:      uuid.New().String(),
		TrackID: trackId,
		FileURL: fileURL,
	}
}
