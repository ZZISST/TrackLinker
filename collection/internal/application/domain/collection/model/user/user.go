package user

import "github.com/google/uuid"

type User struct {
	ID string `json:"id" gorm:"unique;not null;index;primary_key"`
}

func NewAddedContent() User {
	return User{
		ID: uuid.New().String(),
	}
}
