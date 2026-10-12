package service_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
)

func TestFormatLinkForDB(t *testing.T) {
	linkFormatter := modelutils.NewLinkFormater()
	tests := []struct {
		name     string
		input    string
		typeReq  model.MusicRequestType
		expected string
		wantErr  bool
	}{
		{
			"Apple Music Track",
			"https://music.apple.com/us/song/the-beautiful-people/1440747194",
			model.MusicRequestTrack,
			"/song/1440747194",
			false,
		},
		{
			"Spotify Album",
			"https://open.spotify.com/album/7APnZJwayR3lH9aUDPDYQB",
			model.MusicRequestAlbum,
			"/album/7APnZJwayR3lH9aUDPDYQB",
			false,
		},
		{
			"Yandex Track in Album",
			"https://music.yandex.ru/album/2754/track/33267",
			model.MusicRequestTrackInAlbum,
			"/album/2754/track/33267",
			false,
		},
		{
			"Deezer Track",
			"https://www.deezer.com/track/87938845",
			model.MusicRequestTrack,
			"/track/87938845",
			false,
		},
		{
			"Invalid Link",
			"https://example.com/track/12345",
			model.MusicRequestTrack,
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := url.Parse(tt.input)
			assert.NoError(t, err)

			formattedLink, err := linkFormatter.FormatLinkForDB(tt.typeReq, parsedURL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, *formattedLink)
			}
		})
	}
}

func TestFormatLinkFromDBToOutput(t *testing.T) {
	linkFormatter := modelutils.NewLinkFormater()
	tests := []struct {
		name     string
		input    string
		service  model.MusicService
		expected string
		wantErr  bool
	}{
		{
			"Apple Music Track Output",
			"/song/the-beautiful-people/1440747194",
			model.Apple,
			"https://music.apple.com/song/the-beautiful-people/1440747194",
			false,
		},
		{
			"Spotify Album Output",
			"/album/7APnZJwayR3lH9aUDPDYQB",
			model.Spotify,
			"https://open.spotify.com/album/7APnZJwayR3lH9aUDPDYQB",
			false,
		},
		{
			"Invalid Music Service",
			"/track/12345",
			model.Unsupported,
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedLink, err := linkFormatter.FormatLinkFromDBToOutput(tt.input, tt.service)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, *formattedLink)
			}
		})
	}
}
