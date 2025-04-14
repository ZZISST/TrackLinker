/*
Package service provides implementations for music service interactions.

Authors:

	Vladimir Stabrovskiy @godnesty
	Alexander Subbotin @CAHEK_aka_zzisst
	Nikita Kizhvatkin @djsega1
*/
package musicservice

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/config"
)

/*
SpotifyAuthService provides methods for handling Spotify authentication.

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
type SpotifyAuthService struct {
	AccessToken string
	ExpiresAt   time.Time
	cfg         *config.Schema
}

/*
authenticate performs the authentication process with Spotify's API to obtain an access token.

Parameters:

	None

Returns:
  - err (error): An error if the authentication process fails.

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
func (s *SpotifyAuthService) authenticate() error {
	clientID := s.cfg.SpotifyClientID
	clientSecret := s.cfg.SpotifyClientSecret

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

/*
GetValidToken retrieves a valid access token for Spotify's API.
If the current token is expired or invalid, it performs re-authentication.

Parameters:

	None

Returns:
  - token (string): A valid access token for Spotify API requests.
  - err (error): An error if the operation fails.

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
func (s *SpotifyAuthService) GetValidToken() (string, error) {
	if time.Now().After(s.ExpiresAt) {
		if err := s.authenticate(); err != nil {
			return "", err
		}
	}
	return s.AccessToken, nil
}

/*
NewSpotifyAuthService creates and returns a new instance of SpotifyAuthService.

Parameters:

	None

Returns:
  - SpotifyAuthService: A new instance of SpotifyAuthService.

Author: Vladimir Stabrovskiy @godnesty
*/
func NewSpotifyAuthService() SpotifyAuthService {
	return SpotifyAuthService{
		cfg: config.GetConfig(),
	}
}
