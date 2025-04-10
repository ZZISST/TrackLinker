package modelMusic

import "github.com/google/uuid"

type Track struct {
	ID          string `json:"id" gorm:"unique;not null;index;primary_key"`
	Title       string `json:"title"`
	TrackNumber int    `json:"track_number"`

	TrackPreview TrackPreview `json:"track_preview_id" gorm:"foreignKey:TrackID"`
	AlbumID      string       `json:"album_id" gorm:"index"` // Foreign key
	Album        Album        `json:"album" gorm:"constraint:OnDelete:CASCADE;"`
	Artists      []Artist     `json:"artists" gorm:"many2many:track_artists;constraint:OnDelete:CASCADE;"`
	MusicLinks   []MusicLink  `json:"music_links" gorm:"polymorphicType:OwnerType;polymorphicId:OwnerID;polymorphicValue:track"`
}

type TrackPreview struct {
	ID      string `json:"id" gorm:"unique;not null;index;primary_key"`
	FileURL string `json:"file_url"`

	TrackID string `json:"track_id" gorm:"index:idx_tracks_track_preview"`
}

func NewTrack(title string, albumID string) Track {
	return Track{
		ID:      uuid.New().String(),
		Title:   title,
		AlbumID: albumID,
	}
}

func NewTrackPreview(trackId string, fileURL string) TrackPreview {
	return TrackPreview{
		ID:      uuid.New().String(),
		TrackID: trackId,
		FileURL: fileURL,
	}
}
