package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	service "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
)

type MockDeezerMusicService struct{}


func (m *MockDeezerMusicService) GetAllData(title, artist string) (*model.ParsedMusic, error) {
	return service.NewDeezerService().GetAllData(title, artist)
}

func (m *MockDeezerMusicService) GetData(trackUrl string) (string, string, error) {
	return service.NewDeezerService().GetData(trackUrl)
}


func (m *MockDeezerMusicService) Service() model.MusicService {
	return model.Deezer
}

func TestGetAllData_DeezerMusic(t *testing.T) {
	service := &MockDeezerMusicService{}
	result, err := service.GetAllData("Dream of You", "yandere")

	require.NoError(t, err)
	require.NotNil(t, result)
	

	assert.Equal(t, 1043696652, result.ID)
	assert.Equal(t, "Dream of You", result.Title)
	assert.Equal(t, "https://www.deezer.com/track/1043696652", result.Link)
	assert.Equal(t, "yandere", result.Artist.Name)
	assert.Equal(t, "https://www.deezer.com/artist/77843402", result.Artist.Link)
	assert.Equal(t, "https://cdn-images.dzcdn.net/images/cover/97a399885d6aca66d41ffecf1f2b23f7/1000x1000-000000-80-0-0.jpg", result.Album.Cover)
}

func TestGetData_DeezerMusic(t *testing.T) {
	service := &MockDeezerMusicService{}
	title, artist, err := service.GetData("https://www.deezer.com/en/track/1043696652")
	
	assert.NoError(t, err)

	assert.Equal(t, "Dream of You", title)
	assert.Equal(t, "yandere", artist)
}