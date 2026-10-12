package model

import (
	"github.com/google/uuid"
)

type AddedContent struct {
	ID             string `json:"id" gorm:"unique;not null;index;primary_key"`
	OwnerID        string `json:"owner_id" gorm:"index"`
	OwnerType      string `json:"owner_type" gorm:"index"`
	ReferralLinkID string

	CollectionID string      `json:"collection_id" gorm:"index"`
	Listenings   []Listening `json:"listenings" gorm:"foreignKey:AddedContentID"`
}

func NewAddedContent(collectionID, ownerID, ownerType string) AddedContent {
	return AddedContent{
		ID:           uuid.New().String(),
		OwnerID:      ownerID,
		OwnerType:    ownerType,
		CollectionID: collectionID,
	}
}
