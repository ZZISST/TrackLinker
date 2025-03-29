package service_test

import (
	"testing"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	service "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
	"github.com/stretchr/testify/assert"
)

type MockYandexMusicService struct{}

func (m *MockYandexMusicService) GetUrl(title, artist string) (string, error) {
	return service.NewYandexService().GetUrl(title, artist)
}

func (m *MockYandexMusicService) GetData(trackUrl string) (string, string, error) {
	return service.NewYandexService().GetData(trackUrl)
}

func (m *MockYandexMusicService) Service() model.MusicService {
	return model.Yandex
}

func TestGetUrl_YandexMusic(t *testing.T) {
	service := &MockYandexMusicService{}
	url, err := service.GetUrl("Soul Doubt", "$uicideboy$")

	assert.NoError(t, err)
	assert.Equal(t, "https://music.yandex.ru/track/38157634", url)
}

func TestGetData_YandexMusic(t *testing.T) {
	service := &MockYandexMusicService{}
	title, artist, err := service.GetData("https://music.yandex.ru/album/13516170/track/38157634")

	assert.NoError(t, err)
	assert.Equal(t, "Soul Doubt", title)
	assert.Equal(t, "$uicideboy$", artist)
}