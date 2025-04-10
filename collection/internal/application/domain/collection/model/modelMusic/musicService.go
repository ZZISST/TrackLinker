package modelMusic

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
