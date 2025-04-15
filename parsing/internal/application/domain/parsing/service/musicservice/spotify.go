/*
Package service provides implementations for music service interactions.

Authors:

	Vladimir Stabrovskiy @godnesty
	Alexander Subbotin @CAHEK_aka_zzisst
	Nikita Kizhvatkin @djsega1
*/
package musicservice

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	m2 "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/config"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/utils"
)

/*
SpotifyService implements methods to interact with Spotify's music service.

Author: Vladimir Stabrovskiy @godnesty
*/
type SpotifyService struct {
	authService SpotifyAuthService
	cfg         *config.Schema
}

/*
Service returns the music service type represented by SpotifyService.

Parameters:

	None

Returns:
  - model.MusicService: The type of music service (Spotify).

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *SpotifyService) Service() m2.MusicService {
	return m2.Spotify
}

/*
GetTrackData retrieves track metadata from a given Spotify Music URL.

Parameters:
  - trackUrl (string): The URL of the track on Spotify Music.

Returns:
  - title (string): The title of the track.
  - artist (string): The artist of the track.
  - err (error): An error if the operation fails.

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
func (s *SpotifyService) GetTrackData(trackUrl string) (title string, artist string, err error) {
	proxy := s.cfg.ProxyURL
	service := s.authService

	token, err := service.GetValidToken()
	if err != nil {
		return "", "", fmt.Errorf("не удалось получить токен: %v", err)
	}

	rawUrl, _ := url.Parse(trackUrl)
	pathParts := strings.Split(rawUrl.Path, "/")
	for i, part := range pathParts {
		if part == "track" && i+1 < len(pathParts) {
			fmt.Println(pathParts)
		}
	}

	trackID := pathParts[2]

	apiUrl := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", url.PathEscape(trackID))
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	proxyUrl, _ := url.Parse(proxy)

	client := utils.CreateHTTPClient(proxyUrl)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("ошибка Spotify API: %s", body)
	}

	var result struct {
		Name    string `json:"name"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	if len(result.Artists) == 0 {
		return "", "", fmt.Errorf("артист не найден")
	}

	return result.Name, result.Artists[0].Name, nil
}

/*
GetTrackUrl constructs a track URL based on the given title and artist by querying the Spotify API.

Parameters:
  - title (string): The title of the track.
  - artist (string): The name of the artist.

Returns:
  - trackUrl (string): The URL of the track on Spotify.
  - err (error): An error if the operation fails.

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
func (s *SpotifyService) GetTrackUrl(title, artist string) (TrackUrl string, err error) {

	log.Printf("SpotifyService.GetTrackUrl: title: %s, artist: %s", title, artist)

	proxy := s.cfg.ProxyURL
	service := s.authService
	token, err := service.GetValidToken()
	if err != nil {
		return "", fmt.Errorf("не удалось получить токен: %v", err)
	}

	query := url.QueryEscape(fmt.Sprintf("%s %s", title, artist))
	apiUrl := fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=1", query)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	proxyUrl, _ := url.Parse(proxy)

	client := utils.CreateHTTPClient(proxyUrl)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ошибка Spotify API: %s", string(body))
	}

	var result struct {
		Tracks struct {
			Items []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	if len(result.Tracks.Items) == 0 {
		return "", fmt.Errorf("трек не найден")
	}

	trackURL := result.Tracks.Items[0].ExternalUrls.Spotify
	fmt.Println(trackURL)

	return trackURL, nil
}

/*
GetAlbumUrl constructs an album URL based on the given title and artist by querying the Spotify API.

Parameters:
  - title (string): The title of the album.
  - artist (string): The name of the artist.

Returns:
  - trackUrl (string): The URL of the album on Spotify.
  - err (error): An error if the operation fails.

Author:
*/
func (s *SpotifyService) GetAlbumUrl(title, artist string) (trackUrl string, err error) {
	return "", nil
}

/*
GetAlbumData retrieves album metadata from a given Spotify URL.

Parameters:
  - trackUrl (string): The URL of the album on Spotify.

Returns:
  - title (string): The title of the album.
  - artist (string): The artist of the album.
  - err (error): An error if the operation fails.

Author:
*/
func (s *SpotifyService) GetAlbumData(trackUrl string) (title string, artist string, err error) {
	return "", "", nil
}

/*
NewSpotifyService creates and returns a new instance of SpotifyService.

Parameters:

	None

Returns:
  - IParsingMusicService: A new instance of SpotifyService.

Author: Vladimir Stabrovskiy @godnesty
*/
func NewSpotifyService(auth SpotifyAuthService) IParsingMusicService {
	return &SpotifyService{
		authService: auth,
		cfg:         config.GetConfig(),
	}
}
