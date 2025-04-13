package model

import (
	"time"

	"github.com/google/uuid"
)

type Album struct {
	ID          string      `json:"id" gorm:"unique;not null;index;primary_key"`
	Title       string      `json:"title"`
	CoverImage  string      `json:"cover_image"`
	Genres      []Genre     `json:"genres" gorm:"many2many:album_genres;"`
	ReleaseDate time.Time   `json:"release_date"`
	RecordType  string      `json:"record_type"`
	Artists     []Artist    `json:"artists" gorm:"many2many:album_artists;"`
	Type        string      `json:"type"`
	Tracks      []Track     `json:"tracks" gorm:"foreignKey:AlbumID"`
	MusicLinks  []MusicLink `json:"music_links" gorm:"polymorphicType:OwnerType;polymorphicId:OwnerID;polymorphicValue:album"`
}

func NewAlbum(title string, coverImage string, releaseDate time.Time, recordType string) Album {
	return Album{
		ID:          uuid.New().String(),
		Title:       title,
		CoverImage:  coverImage,
		ReleaseDate: releaseDate,
		RecordType:  recordType,
		Type:        "album",
	}
}
