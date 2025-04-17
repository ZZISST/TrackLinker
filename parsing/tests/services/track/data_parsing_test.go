package service_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	service "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
)

// Mock IParsingMusicService
type MockParsingMusicService struct {
	mock.Mock
}

func (m *MockParsingMusicService) GetData(trackUrl string) (string, string, error) {
	args := m.Called(trackUrl)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockParsingMusicService) GetUrl(title, artist string) (string, error) {
	args := m.Called(title, artist)
	return args.String(0), args.Error(1)
}

func (m *MockParsingMusicService) Service() model.MusicService {
	args := m.Called()
	return args.Get(0).(model.MusicService)
}

// Mock IFullParsingMusicService
type MockFullParsingMusicService struct {
	MockParsingMusicService
}

func (m *MockFullParsingMusicService) GetAllData(title, artist string) (*model.ParsedMusic, error) {
	args := m.Called(title, artist)
	return args.Get(0).(*model.ParsedMusic), args.Error(1)
}

func TestGetMusicData(t *testing.T) {
	mockService := new(MockParsingMusicService)
	mockFullService := new(MockFullParsingMusicService)

	parsingService := service.NewParsingService(mockFullService, mockService)

	link, _ := url.Parse("https://music.apple.com/us/song/test-song/123456789")

	mockService.On("Service").Return(model.Apple)
	mockService.On("GetData", link.Path).Return("Test Song", "Test Artist", nil)
	mockService.On("GetUrl", "Test Song", "Test Artist").Return("https://music.apple.com/us/song/test-song/123456789", nil)
	mockFullService.On("GetAllData", "Test Song", "Test Artist").
		Return(&model.ParsedMusic{
			Title:  "Test Song",
			Artist: model.ParsedMusic{}.Artist,
			Album:  model.ParsedMusic{}.Album,
		}, nil)

	links, parsedMusic, err := parsingService.GetMusicDataByLink(link)

	assert.NoError(t, err)
	assert.NotNil(t, links)
	assert.NotNil(t, parsedMusic)
	assert.Equal(t, "Test Song", parsedMusic.Title)
	assert.Equal(t, "Test Artist", parsedMusic.Artist.Name)
}

func TestGetMusicData_Error(t *testing.T) {
	mockService := new(MockParsingMusicService)
	mockFullService := new(MockFullParsingMusicService)
	parsingService := service.NewParsingService(mockFullService, mockService)

	link, _ := url.Parse("https://unsupported.com/song/test-song/123456789")

	_, _, err := parsingService.GetMusicDataByLink(link)
	assert.Error(t, err)
}
