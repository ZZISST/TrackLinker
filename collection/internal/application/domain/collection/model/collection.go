package model

import (
	"github.com/google/uuid"
)

type Collection struct {
	ID            string         `json:"id" gorm:"unique;not null;index;primary_key"`
	UserID        string         `json:"user_id" gorm:"index"`
	AddedContents []AddedContent `json:"added_contents" gorm:"foreignKey:CollectionID"`
}

func NewCollection(userId string) *Collection {
	return &Collection{
		ID: uuid.New().String(),
		UserID: userId,
	}
}
