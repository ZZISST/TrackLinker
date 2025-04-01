package model

import (
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/adapters/db"
)

type MusicServiceDictionary struct {
	Service MusicService `gorm:"primary_key" json:"music_service"`
}

type MusicService string

const (
	Apple       MusicService = "Apple"
	Yandex      MusicService = "Yandex"
	Spotify     MusicService = "Spotify"
	Deezer      MusicService = "Deezer"
	Unsupported MusicService = "Unsupported"
)

type MusicServiceLink string

const (
	AppleCom        MusicServiceLink = "music.apple.com"
	SpotifyCom      MusicServiceLink = "spotify.com"
	OpenSpotifyCom  MusicServiceLink = "open.spotify.com"
	YandexCom       MusicServiceLink = "music.yandex.com"
	YandexRu        MusicServiceLink = "music.yandex.ru"
	DeezerLink      MusicServiceLink = "deezer.page.link"
	DeezerLinkShort MusicServiceLink = "dzr.page.link"
	DeezerCom       MusicServiceLink = "deezer.com"
	UnsupportedLink MusicServiceLink = "Unsupported"
)

func SeedMusicServices(db *db.Adapter) error {
	services := []MusicServiceDictionary{
		{Service: Apple},
		{Service: Yandex},
		{Service: Spotify},
		{Service: Deezer},
		{Service: Unsupported},
	}

	for _, service := range services {
		if err := db.GetDB().FirstOrCreate(&service, service).Error; err != nil {
			return err
		}
	}

	return nil
}
