package service

import (
	"net/url"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

type IBuilderMusic interface {
	setArtists(artistNames []string)
	setGenres(genres []string)
	setAlbum(albumName string)
	setTrack(trackNames []string)
	setTrackPreview(trackPreviewURL string)
	setMusicLink(musicService model.MusicService, url url.URL)
	getMusic()
}

func getBuilder()

type TrackBuilder struct {
}
