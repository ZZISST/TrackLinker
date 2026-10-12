package service_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
)

// MockParsingMusicService мокает musicservice.IParsingMusicService.
type MockParsingMusicService struct {
	mock.Mock
}

func (m *MockParsingMusicService) GetTrackData(trackUrl string) (string, string, error) {
	args := m.Called(trackUrl)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockParsingMusicService) GetAlbumData(albumUrl string) (string, string, error) {
	args := m.Called(albumUrl)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockParsingMusicService) GetTrackUrl(title, artist string) (string, error) {
	args := m.Called(title, artist)
	return args.String(0), args.Error(1)
}

func (m *MockParsingMusicService) GetAlbumUrl(title, artist string) (string, error) {
	args := m.Called(title, artist)
	return args.String(0), args.Error(1)
}

func (m *MockParsingMusicService) Service() model.MusicService {
	return m.Called().Get(0).(model.MusicService)
}

// MockFullParsingMusicService мокает musicservice.IFullParsingMusicService.
type MockFullParsingMusicService struct {
	MockParsingMusicService
}

func (m *MockFullParsingMusicService) GetAllTrackData(title, artist string) (*model.ParsingTrack, error) {
	args := m.Called(title, artist)
	track, _ := args.Get(0).(*model.ParsingTrack)
	return track, args.Error(1)
}

func (m *MockFullParsingMusicService) GetAllAlbumData(title, artist string) (*model.ParsingAlbum, error) {
	args := m.Called(title, artist)
	album, _ := args.Get(0).(*model.ParsingAlbum)
	return album, args.Error(1)
}

const appleTrackLink = "https://music.apple.com/us/song/test-song/123456789"

func TestGetTrackDataByLink(t *testing.T) {
	apple := new(MockParsingMusicService)
	full := new(MockFullParsingMusicService)
	parsingService := service.NewParsingService(full, apple)

	link, err := url.Parse(appleTrackLink)
	require.NoError(t, err)

	apple.On("Service").Return(model.Apple)
	apple.On("GetTrackData", link.Path).Return("Test Song", "Test Artist", nil)
	apple.On("GetTrackUrl", "Test Song", "Test Artist").Return(appleTrackLink, nil)
	full.On("GetAllTrackData", "Test Song", "Test Artist").Return(&model.ParsingTrack{
		Title:   "Test Song",
		Artists: []model.ParsingArtist{{Name: "Test Artist"}},
	}, nil)

	links, track, err := parsingService.GetTrackDataByLink(link)

	require.NoError(t, err)
	require.NotNil(t, links)
	require.NotNil(t, track)
	assert.Equal(t, "Test Song", track.Title)
	assert.Equal(t, "Test Artist", track.Artists[0].Name)
	assert.Contains(t, links.Links, string(model.Apple))
	apple.AssertExpectations(t)
	full.AssertExpectations(t)
}

func TestGetTrackDataByLink_UnsupportedService(t *testing.T) {
	apple := new(MockParsingMusicService)
	full := new(MockFullParsingMusicService)
	parsingService := service.NewParsingService(full, apple)

	link, _ := url.Parse("https://unsupported.com/song/test-song/123456789")

	_, _, err := parsingService.GetTrackDataByLink(link)
	assert.Error(t, err)
	apple.AssertNotCalled(t, "GetTrackData", mock.Anything)
}

func TestGetTrackDataByLink_ParserError(t *testing.T) {
	apple := new(MockParsingMusicService)
	full := new(MockFullParsingMusicService)
	parsingService := service.NewParsingService(full, apple)

	link, _ := url.Parse(appleTrackLink)

	apple.On("Service").Return(model.Apple)
	apple.On("GetTrackData", link.Path).Return("", "", errors.New("api is down"))

	_, _, err := parsingService.GetTrackDataByLink(link)
	assert.Error(t, err)
	full.AssertNotCalled(t, "GetAllTrackData", mock.Anything, mock.Anything)
}

func TestGetTrackDataByData_NilInput(t *testing.T) {
	parsingService := service.NewParsingService(new(MockFullParsingMusicService))

	_, _, err := parsingService.GetTrackDataByData(nil, nil)
	assert.Error(t, err)
}
