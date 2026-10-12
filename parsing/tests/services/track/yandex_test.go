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

func TestGetTrackUrl_Yandex(t *testing.T) {
	svc := musicservice.NewYandexService()
	url, err := svc.GetTrackUrl("Soul Doubt", "$uicideboy$")

	require.NoError(t, err)
	assert.Equal(t, "https://music.yandex.ru/track/38157634", url)
}

func TestGetTrackData_Yandex(t *testing.T) {
	svc := musicservice.NewYandexService()
	title, artist, err := svc.GetTrackData("https://music.yandex.ru/album/13516170/track/38157634")

	require.NoError(t, err)
	assert.Equal(t, "Soul Doubt", title)
	assert.Equal(t, "$uicideboy$", artist)
}
