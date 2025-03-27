package service

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

type YandexService struct {
}

func (s *YandexService) Service() model.MusicService {
	return model.Yandex
}

func (s *YandexService) GetData(trackURL string) (title, artist string, err error) {
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

func (s *YandexService) GetUrl(title, artist string) (trackURL string, err error) {
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

func NewYandexService() IParsingMusicService {
	return &YandexService{}
}
