package model

import "github.com/google/uuid"

type ReferralLink struct {
	ID string `json:"id" gorm:"unique;not null;index;primary_key"`
}

func NewReferralLink() ReferralLink {
	return ReferralLink{
		ID: uuid.New().String(),
	}
}
