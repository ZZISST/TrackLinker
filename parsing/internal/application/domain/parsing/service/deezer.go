package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gocolly/colly"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

type DeezerService struct{}

func (s *DeezerService) Service() model.MusicService {
	return model.Deezer
}

func (s *DeezerService) GetData(trackUrl string) (title string, artist string, err error) {
	c := colly.NewCollector(
		colly.AllowedDomains("deezer.page.link", "dzr.page.link", "www.deezer.com"))

	c.OnHTML("meta[property='og:title']", func(e *colly.HTMLElement) {
		title = e.Attr("content")
	})

	c.OnHTML("meta[itemprop='description']", func(e *colly.HTMLElement) {
		artist = e.Attr("content")
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Ошибка запроса:", err)
	})

	err = c.Visit(trackUrl)
	if err != nil {
		log.Println("Ошибка при посещении страницы:", err)
		return "", "", err
	}

	return title, artist, nil
}

func (s *DeezerService) GetUrl(title, artist string) (TrackUrl string, err error) {
	return "", nil
}

func (s *DeezerService) GetAllData(title, artist string) (*model.ParsedMusic, error) {
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

	var deezerResp dto.DeezerResponse
	if err := json.NewDecoder(resp.Body).Decode(&deezerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deezerResp.Data) == 0 {
		return nil, fmt.Errorf("tracks not find on Deezer")
	}

	firstTrack := deezerResp.Data[0]

	result := model.NewParsedMusic(
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

func NewDeezerService() IFullParsingMusicService {
	return &DeezerService{}
}
