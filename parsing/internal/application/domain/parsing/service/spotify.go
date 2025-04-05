package service

import (
	// "gitlab.com/subbotins11119/course/parsing-service/internal/utils"
	// Spotify "gitlab.com/subbotins11119/course/parsing-service/internal/config/spotify"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	m2 "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/config"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/utils"
)

type SpotifyService struct {
	authService SpotifyAuthService
	cfg         *config.Schema
}

func (s *SpotifyService) Service() m2.MusicService {
	return m2.Spotify
}

func (s *SpotifyService) GetData(trackUrl string) (title string, artist string, err error) {
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

func (s *SpotifyService) GetUrl(title, artist string) (TrackUrl string, err error) {
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

func NewSpotifyService(auth SpotifyAuthService) IParsingMusicService {
	return &SpotifyService{
		authService: auth,
		cfg:         config.GetConfig(),
	}
}
