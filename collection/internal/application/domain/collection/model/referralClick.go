package model

import "github.com/google/uuid"

type ReferralClick struct {
	ID string `json:"id" gorm:"unique;not null;index;primary_key"`
}

func NewReferralClick() ReferralClick {
	return ReferralClick{
		ID: uuid.New().String(),
	}
}
