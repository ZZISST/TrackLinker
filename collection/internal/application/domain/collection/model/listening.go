package model

import (
	"time"

	"github.com/google/uuid"
)

type Listening struct {
	ID             string    `json:"id" gorm:"unique;not null;index;primary_key"`
	TimeStart      time.Time `json:"time_start" gorm:"not null"`
	TimeEnd        time.Time `json:"time_end" gorm:"not null"`
	AddedContentID string    `json:"added_content_id" gorm:"not null;index"`
}

func NewListening(timeStart, timeEnd time.Time, addedContentID string) Listening {
	return Listening{
		ID:             uuid.New().String(),
		TimeStart:      timeStart,
		TimeEnd:        timeEnd,
		AddedContentID: addedContentID,
	}
}
