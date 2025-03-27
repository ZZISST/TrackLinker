package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/utils"
)

// SpotifyAuthService управляет аутентификацией в Spotify API.
type SpotifyAuthService struct {
	AccessToken string
	ExpiresAt   time.Time
}

// authenticate получает новый access token.
func (s *SpotifyAuthService) authenticate() error {
	clientID := "REDACTED"
	clientSecret := "REDACTED"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка получения токена: %s", body)
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	s.AccessToken = authResp.AccessToken
	s.ExpiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)
	return nil
}

// getValidToken возвращает действительный токен, обновляя его при необходимости.
func (s *SpotifyAuthService) GetValidToken() (string, error) {
	if time.Now().After(s.ExpiresAt) {
		if err := s.authenticate(); err != nil {
			return "", err
		}
	}
	return s.AccessToken, nil
}

// GetTrackInfo получает название трека и имя артиста по ID трека.
func GetTrackInfo(trackUrl string) (string, string, error) {
	service := &SpotifyAuthService{}
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

	proxyUrl, _ := url.Parse("http://C13Bo2:M3b6Zo@138.0.241.86:8000")

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

func SpotifyTrackUrl(title, artist string) (string, error) {
	service := &SpotifyAuthService{}
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

	proxyUrl, _ := url.Parse("http://C13Bo2:M3b6Zo@138.0.241.86:8000")

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
	fmt.Println(trackURL) // Для отладки, можно убрать

	return trackURL, nil
}

func NewSpotifyAuthService() SpotifyAuthService {
	return SpotifyAuthService{}
}
