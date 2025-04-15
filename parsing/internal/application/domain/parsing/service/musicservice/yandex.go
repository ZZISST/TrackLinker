/*
Package service provides implementations for music service interactions.

Authors:

	Vladimir Stabrovskiy @godnesty
	Alexander Subbotin @CAHEK_aka_zzisst
	Nikita Kizhvatkin @djsega1
*/
package musicservice

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

/*
AppleService implements methods to interact with Yandex's music service.

Author: Vladimir Stabrovskiy @godnesty
*/
type YandexService struct{}

/*
Service returns the music service type represented by YandexService.

Parameters:

	None

Returns:
  - model.MusicService: The type of music service (Yandex).

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *YandexService) Service() model.MusicService {
	return model.Yandex
}

/*
GetTrackData retrieves track metadata from a given Yandex Music URL.

Parameters:
  - trackUrl (string): The URL of the track on Yandex Music.

Returns:
  - title (string): The title of the track.
  - artist (string): The artist of the track.
  - err (error): An error if the operation fails.

Author:
*/
func (s *YandexService) GetTrackData(trackURL string) (title, artist string, err error) {
	trackID, err := extractTrackID(trackURL)
	if err != nil {
		return "", "", err
	}

	apiURL := fmt.Sprintf("https://music.yandex.ru/handlers/track.jsx?track=%s", trackID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания запроса: %w", err)
	}
	addHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := readResponseBody(resp)
	if err != nil {
		return "", "", err
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return "", "", fmt.Errorf("ошибка разбора JSON: %w", err)
	}

	trackData, ok := jsonResponse["track"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("неверный формат ответа (нет данных о треке)")
	}

	title, _ = trackData["title"].(string)
	artistsData, _ := trackData["artists"].([]interface{})

	var artists []string
	for _, artist := range artistsData {
		if artistMap, ok := artist.(map[string]interface{}); ok {
			if name, exists := artistMap["name"].(string); exists {
				artists = append(artists, name)
			}
		}
	}

	return title, strings.Join(artists, ", "), nil
}

/*
GetTrackUrl constructs a track URL based on the given title and artist by querying the Yandex Music API.

Parameters:
  - title (string): The title of the track.
  - artist (string): The name of the artist.

Returns:
  - trackUrl (string): The URL of the track on Yandex Music.
  - err (error): An error if the operation fails.

Author:
*/
func (s *YandexService) GetTrackUrl(title, artist string) (trackURL string, err error) {

	log.Printf("YandexMusic.GetTrackUrl: title: %s, artist: %s", title, artist)

	searchURL := "https://music.yandex.ru/handlers/music-search.jsx"
	query := fmt.Sprintf("%s %s", title, artist)

	params := url.Values{}
	params.Set("text", query)
	params.Set("type", "all")
	params.Set("lang", "ru")
	params.Set("external-domain", "music.yandex.ru")
	params.Set("overembed", "false")

	req, err := http.NewRequest("GET", searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}
	addHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := readResponseBody(resp)
	if err != nil {
		return "", err
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return "", fmt.Errorf("ошибка разбора JSON: %w", err)
	}

	tracksData, _ := jsonResponse["tracks"].(map[string]interface{})["items"].([]interface{})
	if len(tracksData) == 0 {
		return "", fmt.Errorf("трек не найден")
	}

	firstTrack, _ := tracksData[0].(map[string]interface{})
	trackID, _ := firstTrack["realId"].(string)

	trackUrl := fmt.Sprintf("https://music.yandex.ru/track/%s", trackID)

	if trackID == "" {
		return "", err
	} else {
		return trackUrl, nil
	}

}

/*
GetAlbumUrl constructs an album URL based on the given title and artist by querying the Yandex Music API.

Parameters:
  - title (string): The title of the album.
  - artist (string): The name of the artist.

Returns:
  - trackUrl (string): The URL of the album on Yandex Music.
  - err (error): An error if the operation fails.

Author:
*/
func (s *YandexService) GetAlbumUrl(title, artist string) (trackUrl string, err error) {
	return "", nil
}

/*
GetAlbumData retrieves album metadata from a given Yandex Music URL.

Parameters:
  - trackUrl (string): The URL of the album on Yandex Music.

Returns:
  - title (string): The title of the album.
  - artist (string): The artist of the album.
  - err (error): An error if the operation fails.

Author:
*/
func (s *YandexService) GetAlbumData(trackUrl string) (title string, artist string, err error) {
	return "", "", nil
}

/*
readResponseBody reads and returns the body of an HTTP response.

Parameters:
  - resp (*http.Response): The HTTP response to read the body from.

Returns:
  - []byte: The body of the HTTP response as a byte slice.
  - error: An error if reading the body fails.

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
func readResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("ошибка при разжатии: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	return body, nil
}

/*
addHeaders adds necessary headers to an HTTP request.

Parameters:
  - req (*http.Request): The HTTP request to which headers will be added.

Returns:

	None

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
func addHeaders(req *http.Request) {
	headers := map[string]string{
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 YaBrowser/25.2.0.0 Safari/537.36",
		"Accept":           "application/json, text/javascript, */*; q=0.01",
		"Accept-Encoding":  "gzip, deflate, br, zstd",
		"Accept-Language":  "ru,en;q=0.9",
		"X-Requested-With": "XMLHttpRequest",
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

/*
extractTrackID extracts the track ID from a given Yandex Music track URL.

Parameters:
  - trackURL (string): The URL of the track on Yandex Music.

Returns:
  - string: The extracted track ID.
  - error: An error if the track ID cannot be extracted.

Author: Alexander Subbotin @CAHEK_aka_zzisst
*/
func extractTrackID(trackURL string) (string, error) {
	parsedURL, err := url.Parse(trackURL)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга URL: %w", err)
	}

	pathParts := strings.Split(parsedURL.Path, "/")
	for i, part := range pathParts {
		if part == "track" && i+1 < len(pathParts) {
			return pathParts[i+1], nil
		}
	}

	return "", fmt.Errorf("не удалось извлечь trackID из URL")
}

/*
NewYandexService creates and returns a new instance of YandexService.

Parameters:

	None

Returns:
  - IParsingMusicService: A new instance of YandexService.

Author: Vladimir Stabrovskiy @godnesty
*/
func NewYandexService() IParsingMusicService {
	return &YandexService{}
}
