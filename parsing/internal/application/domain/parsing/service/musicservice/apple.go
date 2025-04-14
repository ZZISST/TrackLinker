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
	"regexp"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/utils"
)

/*
AppleService implements methods to interact with Apple's music service.

Author: Vladimir Stabrovskiy @godnesty
*/
type AppleService struct{}

/*
Service returns the music service type represented by AppleService.

Parameters:

	None

Returns:
  - model.MusicService: The type of music service (Apple).

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *AppleService) Service() model.MusicService {
	return model.Apple
}

/*
GetTrackData retrieves track metadata from a given Apple Music URL.

Parameters:
  - trackUrl (string): The URL of the track on Apple Music.

Returns:
  - title (string): The title of the track.
  - artist (string): The artist of the track.
  - err (error): An error if the operation fails.

Author:
*/
func (s *AppleService) GetTrackData(trackUrl string) (title string, artist string, err error) {

	req, err := http.NewRequest("GET", trackUrl, nil)
	if err != nil {
		return "", "", err
	}

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := utils.CreateHTTPClient(nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", "", err
		}
		defer reader.Close()
	} else {
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", "", err
	}

	bodyStr := string(body)

	titleRegex := regexp.MustCompile(`<meta name="apple:title" content="([^"]+)">`)
	artistRegex := regexp.MustCompile(`<a [^>]*class="click-action[^"]*"[^>]*>([^<]+)</a>`)

	titleMatch := titleRegex.FindStringSubmatch(bodyStr)
	artistMatch := artistRegex.FindStringSubmatch(bodyStr)

	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}
	if len(artistMatch) > 1 {
		artist = artistMatch[1]
	}

	if title == "" || artist == "" {
		return "", "", fmt.Errorf("не удалось получить информацию о треке")
	}

	return title, artist, nil
}

/*
GetTrackUrl constructs a track URL based on the given title and artist by querying the Apple Music API.

Parameters:
  - title (string): The title of the track.
  - artist (string): The name of the artist.

Returns:
  - trackUrl (string): The URL of the track on Apple Music.
  - err (error): An error if the operation fails.

Author:
*/
func (s *AppleService) GetTrackUrl(title, artist string) (trackUrl string, err error) {
	urlStr := "https://amp-api-edge.music.apple.com/v1/catalog/us/search"
	searchTerm := fmt.Sprintf("%s %s", title, artist)

	params := map[string]string{
		"art[music-videos:url]": "c",
		"art[url]":              "f",
		"extend":                "artistUrl",
		"fields[albums]":        "artistName,artistUrl,artwork,contentRating,editorialArtwork,editorialNotes,name,playParams,releaseDate,url,trackCount",
		"fields[artists]":       "url,name,artwork",
		"format[resources]":     "map",
		"include[albums]":       "artists",
		"include[music-videos]": "artists",
		"include[songs]":        "artists",
		"include[stations]":     "radio-show",
		"l":                     "en-US",
		"limit":                 "21",
		"omit[resource]":        "autos",
		"platform":              "web",
		"relate[albums]":        "artists",
		"relate[songs]":         "albums",
		"term":                  searchTerm,
		"types":                 "songs",
		"with":                  "lyricHighlights,lyrics,naturalLanguage,serverBubbles,subtitles",
	}

	queryParams := url.Values{}
	for k, v := range params {
		queryParams.Set(k, v)
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
		return "", err
	}
	u.RawQuery = queryParams.Encode()

	headers := map[string]string{
		"authority":          "amp-api-edge.music.apple.com",
		"method":             "GET",
		"scheme":             "https",
		"accept":             "*/*",
		"accept-encoding":    "gzip, deflate, br, zstd",
		"accept-language":    "ru,en;q=0.9",
		"authorization":      "Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IldlYlBsYXlLaWQifQ.eyJpc3MiOiJBTVBXZWJQbGF5IiwiaWF0IjoxNzM3NDgwNzIxLCJleHAiOjE3NDQ3MzgzMjEsInJvb3RfaHR0cHNfb3JpZ2luIjpbImFwcGxlLmNvbSJdfQ.n_2fV0lbEYcLIKyt590X3A0oH8VFugS53cmTFiGHYxy1ilDRd-rWp9K1Ka2r5aAf-cRdsMaHx7VZSJ4IwG9UnQ",
		"origin":             "https://music.apple.com",
		"priority":           "u=1, i",
		"referer":            "https://music.apple.com/",
		"sec-ch-ua":          `"Not A(Brand";v="8", "Chromium";v="132", "YaBrowser";v="25.2", "Yowser";v="2.5"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 YaBrowser/25.2.0.0 Safari/537.36",
	}

	cookies := map[string]string{
		"s_fid": "53B651DEBC138772-31B6A48B6A3D3DE7",
		"s_vi":  "[CS]v1|33DC60F640C04150-40000AB94479AB03[CE]",
		"geo":   "RU",
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return "", err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	for k, v := range cookies {
		req.AddCookie(&http.Cookie{
			Name:   k,
			Value:  v,
			Domain: ".music.apple.com", // Set the domain
		})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Handle compressed responses
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatalf("Error creating gzip reader: %v", err)
			return "", err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v, body: %s", err, string(body))
		return "", err
	}

	results, ok := data["results"].(map[string]interface{})
	if !ok {
		log.Fatalf("Error extracting results data")
		return "", err
	}

	songData, ok := results["song"].(map[string]interface{})
	if !ok {
		log.Fatalf("Error extracting song data")
		return "", err
	}

	songItems, ok := songData["data"].([]interface{})
	if !ok || len(songItems) == 0 {
		log.Fatalf("Error extracting song items data or no items found")
		return "", err
	}

	firstSong, ok := songItems[0].(map[string]interface{})
	if !ok {
		log.Fatalf("Error extracting first song item")
		return "", err
	}

	trackID, ok := firstSong["id"].(string)
	if !ok {
		log.Fatalf("Error extracting track ID")
		return "", err
	}

	return fmt.Sprintf("https://music.apple.com/us/song/%s", trackID), err
}

/*
GetAlbumUrl constructs an album URL based on the given title and artist by querying the Apple Music API.

Parameters:
  - title (string): The title of the album.
  - artist (string): The name of the artist.

Returns:
  - trackUrl (string): The URL of the album on Apple Music.
  - err (error): An error if the operation fails.

Author:
*/
func (s *AppleService) GetAlbumUrl(title, artist string) (trackUrl string, err error) {
	return "", nil
}

/*
GetAlbumData retrieves album metadata from a given Apple Music URL.

Parameters:
  - trackUrl (string): The URL of the album on Apple Music.

Returns:
  - title (string): The title of the album.
  - artist (string): The artist of the album.
  - err (error): An error if the operation fails.

Author:
*/
func (s *AppleService) GetAlbumData(trackUrl string) (title string, artist string, err error) {
	return "", "", nil
}

/*
NewAppleService creates and returns a new instance of AppleService.

Parameters:

	None

Returns:
  - IParsingMusicService: A new instance of AppleService.

Author: Vladimir Stabrovskiy @godnesty
*/
func NewAppleService() IParsingMusicService {
	return &AppleService{}
}
