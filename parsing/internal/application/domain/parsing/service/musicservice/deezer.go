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
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	_ "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

/*
DeezerService implements methods to interact with Deezer's music service.

Author: Vladimir Stabrovskiy @godnesty
*/
type DeezerService struct{}

/*
Service returns the music service type represented by DeezerService.

Parameters:

	None

Returns:
  - model.MusicService: The type of music service (Deezer).

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *DeezerService) Service() model.MusicService {
	return model.Deezer
}

/*
GetTrackData retrieves track metadata from a given Deezer track URL.

Parameters:
  - trackUrl (string): The URL of the track on Deezer.

Returns:
  - title (string): The title of the track.
  - artist (string): The artist of the track.
  - err (error): An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *DeezerService) GetTrackData(trackUrl string) (title string, artist string, err error) {
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

/*
GetTrackUrl constructs a track URL based on the given title and artist by querying the Deezer API.

Parameters:
  - title (string): The title of the track.
  - artist (string): The name of the artist.

Returns:
  - TrackUrl (string): The URL of the track on Deezer.
  - err (error): An error if the operation fails.

Author:
*/
func (s *DeezerService) GetTrackUrl(title, artist string) (TrackUrl string, err error) {
	return "", nil
}

/*
GetAllTrackData retrieves all metadata for a track based on the given title and artist.

Parameters:
  - title (string): The title of the track.
  - artist (string): The name of the artist.

Returns:
  - *model.ParsingTrack: A struct containing all metadata for the track.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *DeezerService) GetAllTrackData(title, artist string) (*model.ParsingTrack, error) {
	encodedTitle := url.QueryEscape(title)
	encodedArtist := url.QueryEscape(artist)

	searchUrl := fmt.Sprintf("https://api.deezer.com/search?q=%s%%20%s", encodedTitle, encodedArtist)

	req, err := http.NewRequest("GET", searchUrl, nil)
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

	var deezerResp dto.DeezerTrackResponse
	if err := json.NewDecoder(resp.Body).Decode(&deezerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deezerResp.Data) == 0 {
		return nil, fmt.Errorf("tracks not find on Deezer")
	}

	firstTrack := deezerResp.Data[0]

	trackUrl := fmt.Sprintf("https://api.deezer.com/track/%s", firstTrack.ID)

	req, err = http.NewRequest("GET", trackUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a req to Deezer API: %w", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send a req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Deezer API returned status code: %d", resp.StatusCode)
	}

	var deezerTrack dto.DeezerTrack
	if err := json.NewDecoder(resp.Body).Decode(&deezerTrack); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	artists := convertContributorsToParsingArtists(deezerTrack.Contributors)
	album := convertAlbumToParsingAlbum(deezerTrack.Album)

	// Convert release_date to time.Time
	trackReleaseDate, err := time.Parse("2006-01-02", deezerTrack.ReleaseDate)
	if err != nil {
		log.Printf("Failed to parse release date: %v", err)
		trackReleaseDate = time.Time{} // Set to zero value if parsing fails
	}

	// Get genres
	albumUrl := fmt.Sprintf("https://api.deezer.com/album/%s", firstTrack.Album.ID)

	req, err = http.NewRequest("GET", albumUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a req to Deezer API: %w", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send a req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Deezer API returned status code: %d", resp.StatusCode)
	}

	var deezerAlbum dto.DeezerAlbum
	if err := json.NewDecoder(resp.Body).Decode(&deezerAlbum); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// add album release date
	albumReleaseDate, err := time.Parse("2006-01-02", deezerAlbum.ReleaseDate)
	if err != nil {
		log.Printf("Failed to parse release date: %v", err)
		albumReleaseDate = time.Time{} // Set to zero value if parsing fails
	}

	// add album release date
	album.ReleaseDate = albumReleaseDate

	result := model.NewParsingTrack(
		deezerTrack.ID,
		deezerTrack.Title,
		deezerTrack.Duration,
		deezerTrack.Link,
		deezerTrack.TrackPosition,
		deezerTrack.DiscNumber,
		trackReleaseDate,
		artists,
		album,
	)

	return &result, nil
}

/*
GetAllAlbumData retrieves all metadata for an album based on the given title and artist.

Parameters:
  - title (string): The title of the album.
  - artist (string): The name of the artist.

Returns:
  - *model.ParsingAlbum: A struct containing all metadata for the album.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *DeezerService) GetAllAlbumData(title, artist string) (*model.ParsingAlbum, error) {
	encodedTitle := url.QueryEscape(title)
	encodedArtist := url.QueryEscape(artist)

	searchUrl := fmt.Sprintf("https://api.deezer.com/search?q=%s%%20%s", encodedTitle, encodedArtist)

	req, err := http.NewRequest("GET", searchUrl, nil)
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

	var deezerResp dto.DeezerAlbumResponse
	if err := json.NewDecoder(resp.Body).Decode(&deezerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deezerResp.Data) == 0 {
		return nil, fmt.Errorf("tracks not find on Deezer")
	}

	// var searchedAlbum dto.DeezerAlbumnSearch
	var searchedAlbum *dto.DeezerAlbum

	for _, track := range deezerResp.Data {
		if track.Album.Title == title {

			// for Contributors and Genres
			albumUrl := fmt.Sprintf("https://api.deezer.com/album/%s", track.Album.ID)

			req, err = http.NewRequest("GET", albumUrl, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create a req to Deezer API: %w", err)
			}

			resp, err = client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to send a req: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("Deezer API returned status code: %d", resp.StatusCode)
			}

			var deezerAlbum dto.DeezerAlbum
			if err := json.NewDecoder(resp.Body).Decode(&deezerAlbum); err != nil {
				return nil, fmt.Errorf("failed to decode response: %w", err)
			}

			// Check Contributors
			for _, contributor := range deezerAlbum.Contributors {
				if contributor.Name == artist {
					searchedAlbum = &deezerAlbum
					break
				}
			}

			if searchedAlbum != nil {
				// add album release date
				albumReleaseDate, err := time.Parse("2006-01-02", deezerAlbum.ReleaseDate)
				if err != nil {
					log.Printf("Failed to parse release date: %v", err)
					albumReleaseDate = time.Time{} // Set to zero value if parsing fails
				}

				parsingAlbum := convertAlbumToParsingAlbum(*searchedAlbum)
				// add album release date
				parsingAlbum.ReleaseDate = albumReleaseDate
				return &parsingAlbum, nil
			}
		}
	}
	return nil, fmt.Errorf("album not find on Deezer")
}

/*
GetAlbumUrl constructs an album URL based on the given title and artist by querying the Deezer API.

Parameters:
  - title (string): The title of the album.
  - artist (string): The name of the artist.

Returns:
  - trackUrl (string): The URL of the album on Deezer.
  - err (error): An error if the operation fails.

Author:
*/
func (s *DeezerService) GetAlbumUrl(title, artist string) (trackUrl string, err error) {
	return "", nil
}

/*
GetAlbumData retrieves album metadata from a given Deezer album URL.

Parameters:
  - trackUrl (string): The URL of the album on Deezer.

Returns:
  - title (string): The title of the album.
  - artist (string): The artist of the album.
  - err (error): An error if the operation fails.

Author:
*/
func (s *DeezerService) GetAlbumData(trackUrl string) (title string, artist string, err error) {
	return "", "", nil
}

/*
deezerExtractTrackID extracts the track ID from a given Deezer track URL.

Parameters:
  - trackUrl (string): The URL of the track on Deezer.

Returns:
  - trackID (string): The extracted track ID.

Author: Vladimir Stabrovskiy @godnesty
*/
func deezerExtractTrackID(trackUrl string) string {
	re := regexp.MustCompile(`(?:track|x).*?(\d+)`)

	matches := re.FindStringSubmatch(trackUrl)

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func convertContributorsToParsingArtists(contributors []dto.DeezerContributor) []model.ParsingArtist {
	var artists []model.ParsingArtist
	for _, contributor := range contributors {
		artists = append(artists, model.ParsingArtist{
			ID:    contributor.ID,
			Name:  contributor.Name,
			Cover: contributor.PictureXl,
		})
	}
	return artists
}

// Convert deezerTrack.Album to ParsingAlbum
func convertAlbumToParsingAlbum(album dto.DeezerAlbum) model.ParsingAlbum {
	return model.ParsingAlbum{
		ID:         album.ID,
		Title:      album.Title,
		Cover:      album.Cover,
		RecordType: album.RecordType,
		Genres:     convertGenresToParsingGenres(album.Genres),
	}
}

func convertGenresToParsingGenres(parsedGenres []dto.DeezerGenre) []model.ParsingGenre {
	var genres []model.ParsingGenre
	for _, genre := range parsedGenres {
		genres = append(genres, model.ParsingGenre{
			ID:   genre.ID,
			Name: genre.Name,
		})
	}
	return genres
}

func NewDeezerService() IFullParsingMusicService {
	return &DeezerService{}
}
