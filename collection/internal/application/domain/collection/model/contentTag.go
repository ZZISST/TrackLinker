package model

import "github.com/google/uuid"

type ContentTag struct {
	ID string `json:"id" gorm:"unique;not null;index;primary_key"`
}

func NewContentTag() ContentTag {
	return ContentTag{
		ID: uuid.New().String(),
	}
}
