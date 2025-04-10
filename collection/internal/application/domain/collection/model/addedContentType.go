package model

import (
	"github.com/google/uuid"
)

type AddedContentType struct {
	ID string `json:"id" gorm:"unique;not null;index;primary_key"`
}

func NewAddedContentType() AddedContentType {
	return AddedContentType{
		ID: uuid.New().String(),
	}
}
