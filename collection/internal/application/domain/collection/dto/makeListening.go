package dto

import "time"

type RequestMakeListening struct {
	UserID    string
	Name      string
	Artists   string
	TimeStart time.Time
	TimeEnd   time.Time
}

type ResponseMakeListening struct {
	Status string `json:"status"`
}
