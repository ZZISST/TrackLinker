package model

import (
	"time"

	"github.com/google/uuid"
)

type Album struct {
	ID         string    `json:"id" gorm:"unique;not null;index;primary_key"`
	Title      string    `json:"title"`
	ArtistID   string    `json:"artist_id"`
	CoverImage string    `json:"cover_image"`
	ReleasedAt time.Time `json:"released_at"`
	CreatedAt  time.Time `json:"created_at"`

	Artists    []Artist    `json:"artists" gorm:"many2many:album_artists;"`
	Tracks     []Track     `json:"tracks" gorm:"foreignKey:AlbumID"`
	MusicLinks []MusicLink `json:"music_links" gorm:"polymorphicType:OwnerType;polymorphicId:OwnerID;polymorphicValue:album"`
}

func NewAlbum(title string, artistId string, coverImage string, releasedAt time.Time) Album {
	return Album{
		ID:         uuid.New().String(),
		Title:      title,
		ArtistID:   artistId,
		CoverImage: coverImage,
		ReleasedAt: releasedAt,
		CreatedAt:  time.Now(),
	}
}
