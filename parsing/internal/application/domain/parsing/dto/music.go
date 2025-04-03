package dto

import (
	"encoding/json"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

type MusicTrack struct {
	TrackID      string             `json:"track_id" gorm:"-"`
	Album        model.Album        `json:"album" gorm:"-"`
	Name         string             `json:"name" gorm:"-"`
	TrackNumber  int                `json:"track_number" gorm:"-"`
	Artists      []model.Artist     `json:"artists" gorm:"-"`
	Links        []model.MusicLink  `json:"links" gorm:"-"`
	TrackPreview model.TrackPreview `json:"track_preview" gorm:"-"`
}

type MusicResultDto struct {
	Type string          `json:"type" gorm:"-"`                      // "track", "album", "playlist"
	Data json.RawMessage `json:"data" gorm:"-" swaggertype:"object"` // в зависимости от типа здесь будет другой JSON
}

type MusicAlbum struct {
	AlbumID    string    `json:"album_id" gorm:"-"`
	Name       string    `json:"name" gorm:"-"`
	CoverImage string    `json:"cover_image"`
	ReleasedAt time.Time `json:"released_at"`
	CreatedAt  time.Time `json:"created_at"`

	Artists []model.Artist    `json:"artists" gorm:"-" swaggertype:"array,string"`
	Links   []model.MusicLink `json:"links" gorm:"-" swaggertype:"array,string"`
	Tracks  []model.Track     `json:"tracks" gorm:"-" swaggertype:"array,string"`
}

type FindMusicReq struct {
	Link string `json:"link" gorm:"-"`
}
