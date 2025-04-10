package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type ParsedMusic struct {
	ID      int    `json:"id" gorm:"-"`
	Title   string `json:"title" gorm:"-"`
	Link    string `json:"link" gorm:"-"`
	Preview string `json:"preview" gorm:"-"`
	Artist  struct {
		Name string `json:"name" gorm:"-"`
		Link string `json:"link" gorm:"-"`
	} `json:"artist" gorm:"-"`
	Album struct {
		Cover string `json:"cover_xl" gorm:"-"`
	} `json:"album" gorm:"-"`
}

type DeezerSong struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Preview string `json:"preview"`
	Artist  struct {
		Name string `json:"name"`
		Link string `json:"link"`
	} `json:"artist"`
	Album struct {
		Cover string `json:"cover_xl"`
	} `json:"album"`
}

type DeezerResponse struct {
	Data []DeezerSong `json:"data"`
}

func NewParsedMusic(id int, title string, link string, preview string, artistName string, artistLink string, albumCover string) ParsedMusic {
	return ParsedMusic{
		ID:      id,
		Title:   title,
		Link:    link,
		Preview: preview,
		Artist: struct {
			Name string `json:"name" gorm:"-"`
			Link string `json:"link" gorm:"-"`
		}{
			Name: artistName,
			Link: artistLink,
		},
		Album: struct {
			Cover string `json:"cover_xl" gorm:"-"`
		}{
			Cover: albumCover,
		},
	}
}

type MusicService string

const (
	Apple       MusicService = "Apple"
	Yandex      MusicService = "Yandex"
	Spotify     MusicService = "Spotify"
	Deezer      MusicService = "Deezer"
	Unsupported MusicService = "Unsupported"
)

type DeezerService struct{}

func (s *DeezerService) Service() MusicService {
	return Deezer
}

func (s *DeezerService) GetData(trackUrl string) (title string, artist string, err error) {
	trackID := deezerExtractTrackID(trackUrl)
	if trackID == "" {
		return "", "", fmt.Errorf("не удалось извлечь ID трека из URL")
	}

	apiURL := fmt.Sprintf("https://api.deezer.com/track/%s", trackID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", fmt.Errorf("ошибка выполнения запроса к API Deezer: %w", err)
	}
	defer resp.Body.Close()

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return "", "", fmt.Errorf("ошибка разбора ответа от API Deezer: %w", err)
	}

	title, ok := jsonResponse["title"].(string)
	if !ok {
		return "", "", fmt.Errorf(apiURL)
	}

	artistData, ok := jsonResponse["artist"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("не удалось извлечь данные об исполнител")
	}

	artist, ok = artistData["name"].(string)
	if !ok {
		return "", "", fmt.Errorf("не удалось извлечь имя исполнителя из ответа")
	}

	return title, artist, nil
}

func (s *DeezerService) GetUrl(title, artist string) (TrackUrl string, err error) {
	return "", nil
}

func (s *DeezerService) GetAllData(title, artist string) (*ParsedMusic, error) {
	encodedTitle := url.QueryEscape(title)
	encodedArtist := url.QueryEscape(artist)

	apiUrl := fmt.Sprintf("https://api.deezer.com/search?q=%s%%20%s", encodedTitle, encodedArtist)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a req to Deezer API: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send a req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Deezer API returned status code: %d", resp.StatusCode)
	}

	var deezerResp DeezerResponse
	if err := json.NewDecoder(resp.Body).Decode(&deezerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deezerResp.Data) == 0 {
		return nil, fmt.Errorf("tracks not find on Deezer")
	}

	firstTrack := deezerResp.Data[0]

	result := NewParsedMusic(
		firstTrack.ID,
		firstTrack.Title,
		firstTrack.Link,
		firstTrack.Preview,
		firstTrack.Artist.Name,
		firstTrack.Artist.Link,
		firstTrack.Album.Cover,
	)

	return &result, nil
}

func deezerExtractTrackID(trackUrl string) string {
	re := regexp.MustCompile(`(?:track|x).*?(\d+)`)

	matches := re.FindStringSubmatch(trackUrl)

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
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


func YanexGet(trackURL string) (title, artist string, err error){
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
		return "", "", fmt.Errorf("неверный формат ответа (капча))")
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


func GetUrl(title, artist string) (trackURL string, err error) {
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

func main() {
	// service := &DeezerService{}

	// title, artist, err := service.GetData("https://www.deezer.com/track/3135556")
	// if err != nil {
	// 	fmt.Println("Ошибка GetData:", err)
	// } else {
	// 	fmt.Printf("Название: %s, Артист: %s\n", title, artist)
	// }

	// musicData, err := service.GetAllData("Harder Better Faster Stronger", "Daft Punk")
	// if err != nil {
	// 	fmt.Println("Ошибка GetAllData:", err)
	// } else {
	// 	fmt.Printf("Результат GetAllData: %+v\n", *musicData)
	// }


	// fmt.Printf(YanexGet("https://music.yandex.ru/album/13516170/track/38157634"))
	fmt.Printf(GetUrl("Soul Doubt", "$uicideboy$"))
}
