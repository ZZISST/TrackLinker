package model

import (
	"github.com/google/uuid"
)

type Artist struct {
	ID     string `json:"id" gorm:"unique;not null;index;primary_key"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`

	Albums []Album `json:"albums" gorm:"many2many:album_artists;"`
}

func NewArtist(name string, avatar string) Artist {
	return Artist{
		ID:     uuid.New().String(),
		Name:   name,
		Avatar: avatar,
	}
}
