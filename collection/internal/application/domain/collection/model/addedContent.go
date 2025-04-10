package model

import (
	"github.com/google/uuid"
)

type AddedContent struct {
	ID             string `json:"id" gorm:"unique;not null;index;primary_key"`
	UserID         string
	OwnerID        string `json:"owner_id" gorm:"index"`
	OwnerType      string `json:"owner_type" gorm:"index"`
	ReferralLinkID string

	Listenings []Listening `json:"listenings" gorm:"foreignKey:ListeningID"`
}

func NewAddedContent(userID, ownerID, ownerType string) AddedContent {
	return AddedContent{
		ID:        uuid.New().String(),
		UserID:    userID,
		OwnerID:   ownerID,
		OwnerType: ownerType,
	}
}
