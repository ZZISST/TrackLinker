package dto

import "encoding/json"

type MusicResult struct {
	Data json.RawMessage `json:"data" gorm:"-" swaggertype:"object"`
}

type FindMusicReq struct {
	Link string `json:"link" gorm:"-"`
}
