package modelMusic

import "github.com/google/uuid"

type MusicLink struct {
	ID      string       `json:"id" gorm:"unique;not null;index;primary_key"`
	Service MusicService `json:"music_service"`
	URL     string       `json:"url"`

	// One music link belongs to one track
	OwnerID   string `json:"owner_id" gorm:"index"`
	OwnerType string `json:"owner_type" gorm:"index"`
}

type OwnerType string

const (
	OwnerTypeTrack OwnerType = "track"
	OwnerTypeAlbum OwnerType = "album"
)

func NewMusicLink(service MusicService, url string, ownerID string, ownerType string) MusicLink {
	return MusicLink{
		ID:        uuid.New().String(),
		Service:   service,
		URL:       url,
		OwnerID:   ownerID,
		OwnerType: ownerType,
	}
}
