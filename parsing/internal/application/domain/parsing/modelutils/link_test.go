package modelutils

import (
	"net/url"
	"testing"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

func TestIdentifyMusicServiceByLink(t *testing.T) {
	linkFormater := NewLinkFormater()

	tests := []struct {
		name          string
		inputURL      string
		expected      model.MusicService
		expectedError bool
	}{
		{"Apple Music", "https://music.apple.com/us/song/the-beautiful-people/1440747194", model.Apple, false},
		{"Yandex Music", "https://music.yandex.ru/album/2754/track/33267", model.Yandex, false},
		{"Spotify", "https://open.spotify.com/track/2aIB1CdRRG7YLBu9nR", model.Spotify, false},
		{"Deezer", "https://www.deezer.com/track/87938845", model.Deezer, false},
		{"Unsupported Service", "https://unsupported.service.com/track/12345", model.Unsupported, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, _ := url.Parse(tt.inputURL)
			service, err := linkFormater.IdentifyMusicServiceByLink(parsedURL)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if service != tt.expected {
				t.Errorf("expected service: %v, got: %v", tt.expected, service)
			}
		})
	}
}

func TestIdentifyMusicTypeByLink(t *testing.T) {
	linkFormater := NewLinkFormater()

	tests := []struct {
		name          string
		inputURL      string
		expected      model.MusicRequestType
		expectedError bool
	}{
		{"Apple Track", "https://music.apple.com/us/song/the-beautiful-people/1440747194", model.MusicRequestTrack, false},
		{"Apple Album", "https://music.apple.com/us/album/lest-we-forget-the-best-of-marilyn-manson/1440746922", model.MusicRequestAlbum, false},
		{"Yandex Track in Album", "https://music.yandex.ru/album/2754/track/33267", model.MusicRequestTrackInAlbum, false},
		{"Yandex Album", "https://music.yandex.ru/album/2754", model.MusicRequestAlbum, false},
		{"Yandex Track", "https://music.yandex.ru/track/111402809", model.MusicRequestTrack, false},
		{"Spotify Track", "https://open.spotify.com/track/2aIB1CdRRG7YLBu9nR", model.MusicRequestTrack, false},
		{"Spotify Album", "https://open.spotify.com/album/7APnZJwayR3lH9aUDPDYQB", model.MusicRequestAlbum, false},
		{"Deezer Track", "https://www.deezer.com/track/87938845", model.MusicRequestTrack, false},
		{"Deezer Album", "https://www.deezer.com/album/8880231", model.MusicRequestAlbum, false},
		{"Unsupported Link", "https://unsupported.service.com/track/12345", model.MusicRequestType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, _ := url.Parse(tt.inputURL)
			musicType, err := linkFormater.IdentifyMusicTypeByLink(parsedURL)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if musicType != nil && *musicType != tt.expected {
				t.Errorf("expected music type: %v, got: %v", tt.expected, *musicType)
			}
		})
	}
}

func TestFormatLinkForDB(t *testing.T) {
	linkFormater := NewLinkFormater()

	tests := []struct {
		name          string
		inputURL      string
		mtype         model.MusicRequestType
		expected      string
		expectedError bool
	}{
		{"Apple Track With Region and Name", "https://music.apple.com/us/song/the-beautiful-people/1440747194", model.MusicRequestTrack, "/song/1440747194", false},
		{"Apple Album With Region and Name", "https://music.apple.com/us/album/lest-we-forget-the-best-of-marilyn-manson/1440746922", model.MusicRequestAlbum, "/album/1440746922", false},
		{"Apple Track With No Region and Name", "https://music.apple.com/song/the-beautiful-people/1440747194", model.MusicRequestTrack, "/song/1440747194", false},
		{"Apple Album With No Region and Name", "https://music.apple.com/album/lest-we-forget-the-best-of-marilyn-manson/1440746922", model.MusicRequestAlbum, "/album/1440746922", false},
		{"Apple Track With Region and No Name", "https://music.apple.com/us/song/1440747194", model.MusicRequestTrack, "/song/1440747194", false},
		{"Apple Album With Region and No Name", "https://music.apple.com/us/album/1440746922", model.MusicRequestAlbum, "/album/1440746922", false},
		{"Apple Track With No Region and No Name", "https://music.apple.com/song/1440747194", model.MusicRequestTrack, "/song/1440747194", false},
		{"Apple Album With No Region and No Name", "https://music.apple.com/album/1440746922", model.MusicRequestAlbum, "/album/1440746922", false},
		{"Yandex Track", "https://music.yandex.ru/album/2754/track/33267", model.MusicRequestTrackInAlbum, "/album/2754/track/33267", false},
		{"Yandex Album", "https://music.yandex.ru/album/2754", model.MusicRequestAlbum, "/album/2754", false},
		{"Yandex Track in Album", "https://music.yandex.ru/album/2754/track/33267", model.MusicRequestTrackInAlbum, "/album/2754/track/33267", false},
		{"Spotify Track", "https://open.spotify.com/track/2aIB1CdRRG7YLBu9nR", model.MusicRequestTrack, "/track/2aIB1CdRRG7YLBu9nR", false},
		{"Spotify Album", "https://open.spotify.com/album/7APnZJwayR3lH9aUDPDYQB", model.MusicRequestAlbum, "/album/7APnZJwayR3lH9aUDPDYQB", false},
		{"Deezer Album", "https://www.deezer.com/album/8880231", model.MusicRequestAlbum, "/album/8880231", false},
		{"Deezer Track", "https://www.deezer.com/track/87938845", model.MusicRequestTrack, "/track/87938845", false},
		{"Unsupported Link", "https://unsupported.service.com/track/12345", model.MusicRequestTrack, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, _ := url.Parse(tt.inputURL)
			formattedLink, err := linkFormater.FormatLinkForDB(tt.mtype, parsedURL)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if formattedLink != nil && *formattedLink != tt.expected {
				t.Errorf("expected formatted link: %v, got: %v", tt.expected, *formattedLink)
			}
		})
	}
}

func TestFormatLinkFromDBToOutput(t *testing.T) {
	linkFormater := NewLinkFormater()

	tests := []struct {
		name          string
		dbFormat      string
		musicService  model.MusicService
		expected      string
		expectedError bool
	}{
		{"Apple Track", "/song/the-beautiful-people/1440747194", model.Apple, "https://music.apple.com/song/the-beautiful-people/1440747194", false},
		{"Apple Album", "/album/lest-we-forget-the-best-of-marilyn-manson/1440746922", model.Apple, "https://music.apple.com/album/lest-we-forget-the-best-of-marilyn-manson/1440746922", false},
		{"Yandex Track", "/track/33267", model.Yandex, "https://music.yandex.com/track/33267", false},
		{"Yandex Track in Album", "/album/2754/track/33267", model.Yandex, "https://music.yandex.com/album/2754/track/33267", false},
		{"Yandex Album", "/album/2754", model.Yandex, "https://music.yandex.com/album/2754", false},
		{"Spotify Track", "/track/2aIB1CdRRG7YLBu9nR", model.Spotify, "https://open.spotify.com/track/2aIB1CdRRG7YLBu9nR", false},
		{"Spotify Album", "/album/7APnZJwayR3lH9aUDPDYQB", model.Spotify, "https://open.spotify.com/album/7APnZJwayR3lH9aUDPDYQB", false},
		{"Deezer Album", "/album/8880231", model.Deezer, "https://deezer.page.link/album/8880231", false},
		{"Deezer Track", "/track/87938845", model.Deezer, "https://deezer.page.link/track/87938845", false},
		{"Unsupported Service", "/track/12345", model.Unsupported, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedLink, err := linkFormater.FormatLinkFromDBToOutput(tt.dbFormat, tt.musicService)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if formattedLink != nil && *formattedLink != tt.expected {
				t.Errorf("expected formatted link: %v, got: %v", tt.expected, *formattedLink)
			}
		})
	}
}
