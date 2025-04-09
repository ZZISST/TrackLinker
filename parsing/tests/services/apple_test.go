package service_test

import (
	"testing"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	service "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
	"github.com/stretchr/testify/assert"
)

type MockAppleMusicService struct{}

func (m *MockAppleMusicService) GetUrl(title, artist string) (string, error) {
	return service.NewAppleService().GetUrl(title, artist)
}
// TODO: починить этот метод (возвращает не верного артиста)
func (m *MockAppleMusicService) GetData(trackUrl string) (string, string, error) {
	return service.NewAppleService().GetData(trackUrl)
}

func (m *MockAppleMusicService) Service() model.MusicService {
	return model.Apple
}

func TestGetUrl_AppleMusic(t *testing.T) {
	service := &MockAppleMusicService{}
	url, err := service.GetUrl("Still Don't Know My Name", "Labrinth")

	assert.NoError(t, err)
	assert.Equal(t, "https://music.apple.com/us/song/1481092695", url)
}

func TestGetData_AppleMusic(t *testing.T) {
	service := &MockAppleMusicService{}
	title, artist, err := service.GetData("https://music.apple.com/ru/album/still-dont-know-my-name/1481092579?i=1481092695")

	assert.NoError(t, err)
	assert.Equal(t, "Still Don't Know My Name", title)
	assert.Equal(t, "Labrinth", artist)
}


