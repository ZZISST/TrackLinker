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

func TestGetTrackUrl_Apple(t *testing.T) {
	svc := musicservice.NewAppleService()
	url, err := svc.GetTrackUrl("Still Don't Know My Name", "Labrinth")

	require.NoError(t, err)
	assert.Equal(t, "https://music.apple.com/us/song/1481092695", url)
}

func TestGetTrackData_Apple(t *testing.T) {
	svc := musicservice.NewAppleService()
	title, artist, err := svc.GetTrackData("https://music.apple.com/ru/album/still-dont-know-my-name/1481092579?i=1481092695")

	require.NoError(t, err)
	assert.Equal(t, "Still Don't Know My Name", title)
	assert.Equal(t, "Labrinth", artist)
}
