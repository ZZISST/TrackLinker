package model

import "github.com/google/uuid"

type Genre struct {
	ID   string `json:"id" gorm:"unique;not null;index;primary_key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func NewGenre(name string) Genre {
	return Genre{
		ID:   uuid.New().String(),
		Name: name,
		Type: "genre",
	}
}
