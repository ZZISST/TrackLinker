package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)



func GetAllAlbumData(title, artist string) (*model.ParsingAlbum, error) {
	encodedTitle := url.QueryEscape(title)
	encodedArtist := url.QueryEscape(artist)

	searchUrl := fmt.Sprintf("https://api.deezer.com/search/album?q=%s%%20%s", encodedTitle, encodedArtist)

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
	
	
	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read response body: %w", err)
	// }
	// fmt.Println(string(bodyBytes))

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
		albumUrl := fmt.Sprintf("https://api.deezer.com/album/%d", track.ID)

		req, err := http.NewRequest("GET", albumUrl, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create a req to Deezer API: %w", err)
		}

		resp, err := client.Do(req)
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
		jsonDump, _ := json.MarshalIndent(deezerAlbum, "", "  ")
		log.Printf("Decoded Album JSON:\n%s", string(jsonDump))
		
		// Проверка, содержит ли Contributors нужного артиста
		for _, contributor := range deezerAlbum.Contributors {
			if contributor.Name == artist {
				searchedAlbum = &deezerAlbum
				break
			}
		}

		if searchedAlbum != nil {
			// Добавляем Contributors к каждому треку (они в основном одинаковые с альбомом)
			for i := range searchedAlbum.Tracks.Data {
				searchedAlbum.Tracks.Data[i].Contributors = searchedAlbum.Contributors
				searchedAlbum.Tracks.Data[i].Album = *searchedAlbum // заполняем Album в каждом треке
				searchedAlbum.Tracks.Data[i].ReleaseDate = searchedAlbum.ReleaseDate
			}

			// Парсим дату
			albumReleaseDate, err := time.Parse("2006-01-02", searchedAlbum.ReleaseDate)
			if err != nil {
				log.Printf("Failed to parse release date: %v", err)
				albumReleaseDate = time.Time{}
			}

			parsingAlbum := convertAlbumToParsingAlbum(*searchedAlbum)
			parsingAlbum.ReleaseDate = albumReleaseDate
			return &parsingAlbum, nil
		}
	}

	return nil, fmt.Errorf("album not found on Deezer")	
}



func convertAlbumToParsingAlbum(album dto.DeezerAlbum) model.ParsingAlbum {
	return model.ParsingAlbum{
		ID:         album.ID,
		Title:      album.Title,
		Cover:      album.Cover,
		RecordType: album.RecordType,
		Genres:     convertGenresToParsingGenres(album.Genres.Data),
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

func main() {

	fmt.Println(GetAllAlbumData("dj billybool", "BLOMSTERTID"))

}