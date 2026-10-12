//go:build integration

// Интеграционный тест: ходит в реальный API.
// Запуск: go test -tags integration ./tests/...
package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service/musicservice"
)

func TestGetTrackUrl_Spotify(t *testing.T) {
	svc := musicservice.NewSpotifyService(musicservice.NewSpotifyAuthService())
	url, err := svc.GetTrackUrl("Romeo", "Bladee")

	require.NoError(t, err)
	assert.Equal(t, "https://open.spotify.com/track/2KimUZVFyJ5oK4d10Y1RUp", url)
}

func TestGetTrackData_Spotify(t *testing.T) {
	svc := musicservice.NewSpotifyService(musicservice.NewSpotifyAuthService())
	title, artist, err := svc.GetTrackData("https://open.spotify.com/track/2KimUZVFyJ5oK4d10Y1RUp")

	require.NoError(t, err)
	assert.Equal(t, "Romeo", title)
	assert.Equal(t, "Bladee", artist)
}
