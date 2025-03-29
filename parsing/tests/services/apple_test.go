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

func (m *MockAppleMusicService) GetData(trackUrl string) (string, string, error) {
	return service.NewAppleService().GetData(trackUrl)
}

func (m *MockAppleMusicService) Service() model.MusicService {
	return model.Apple
}

func TestGetUrl_AppleMusic(t *testing.T) {
	service := &MockAppleMusicService{}
	url, err := service.GetUrl("Big Ballaa (feat. Tezzus)", "Brennan Jones")

	assert.NoError(t, err)
	assert.Equal(t, "https://music.apple.com/us/song/1795507401", url)
}

func TestGetData_AppleMusic(t *testing.T) {
	service := &MockAppleMusicService{}
	title, artist, err := service.GetData("https://music.apple.com/us/song/1795507401")

	assert.NoError(t, err)
	assert.Equal(t, "Big Ballaa (feat. Tezzus)", title)
	assert.Equal(t, "Brennan Jones", artist)
}


