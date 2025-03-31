	package service_test

	import (
		"testing"

		"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
		service "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
		"github.com/stretchr/testify/assert"
	)

	type MockSpotifyMusicService struct{}

	func (m *MockSpotifyMusicService) GetUrl(title, artist string) (string, error) {
		return service.NewSpotifyService(service.NewSpotifyAuthService()).GetUrl(title, artist)
	}

	func (m *MockSpotifyMusicService) GetData(trackUrl string) (string, string, error) {
		return service.NewSpotifyService(service.NewSpotifyAuthService()).GetData(trackUrl)
	}

	func (m *MockSpotifyMusicService) Service() model.MusicService {
		return model.Spotify
	}

	func TestGetUrl_SpotifyMusic(t *testing.T) {
		service := &MockSpotifyMusicService{}
		url, err := service.GetUrl("Romeo", "Bladee")

		assert.NoError(t, err)
		assert.Equal(t, "https://open.spotify.com/track/2KimUZVFyJ5oK4d10Y1RUp", url)
	}

	func TestGetData_SpotifyMusic(t *testing.T) {
		service := &MockSpotifyMusicService{}
		title, artist, err := service.GetData("https://open.spotify.com/track/2KimUZVFyJ5oK4d10Y1RUp")

		assert.NoError(t, err)
		assert.Equal(t, "Romeo", title)
		assert.Equal(t, "Bladee", artist)
	}