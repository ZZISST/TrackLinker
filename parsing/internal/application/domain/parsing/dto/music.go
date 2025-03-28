package dto

import "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"

type Music struct {
	TrackID      string             `json:"track_id" gorm:"-"`
	Album        model.Album        `json:"album" gorm:"-"`
	Name         string             `json:"name" gorm:"-"`
	Artists      []model.Artist     `json:"artists" gorm:"-"`
	Links        []model.MusicLink  `json:"links" gorm:"-"`
	TrackPreview model.TrackPreview `json:"track_preview" gorm:"-"`
}

type FindMusicReq struct {
	Link string `json:"link" gorm:"-"`
}
