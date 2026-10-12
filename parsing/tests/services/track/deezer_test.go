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

func TestGetAllTrackData_Deezer(t *testing.T) {
	svc := musicservice.NewDeezerService()
	track, err := svc.GetAllTrackData("Dream of You", "yandere")

	require.NoError(t, err)
	require.NotNil(t, track)
	assert.Equal(t, 1043696652, track.ID)
	assert.Equal(t, "Dream of You", track.Title)
	require.NotEmpty(t, track.Artists)
	assert.Equal(t, "yandere", track.Artists[0].Name)
}

func TestGetTrackData_Deezer(t *testing.T) {
	svc := musicservice.NewDeezerService()
	title, artist, err := svc.GetTrackData("https://www.deezer.com/en/track/1043696652")

	require.NoError(t, err)
	assert.Equal(t, "Dream of You", title)
	assert.Equal(t, "yandere", artist)
}
