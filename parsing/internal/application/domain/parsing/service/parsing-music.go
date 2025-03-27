package service

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

type IParsingMusicService interface {
	GetData(trackUrl string) (title string, artist string, err error)
	GetUrl(title, artist string) (trackUrl string, err error)
	Service() model.MusicService
}

type IFullParsingMusicService interface {
	IParsingMusicService
	GetAllData(title, artist string) (*model.ParsedMusic, error)
}

type IParsingService interface {
	getLinks(title, artist string) (*model.LinkData, error)
	identifyMusicServiceByLink(link *url.URL) (model.MusicService, error)
	getAllData(title, artist string) (*model.ParsedMusic, error)
	getParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error)
	GetMusicData(link *url.URL) (*model.LinkData, *model.ParsedMusic, error)
}

type ParsingService struct {
	services   []IParsingMusicService
	fullParser IFullParsingMusicService
}

func (p *ParsingService) identifyMusicServiceByLink(link *url.URL) (model.MusicService, error) {
	var streamingService model.MusicService

	host := link.Host

	// identify music service
	switch {
	case strings.Contains(host, string(model.AppleCom)):
		streamingService = model.Apple
	case strings.Contains(host, string(model.YandexCom)) || strings.Contains(host, string(model.YandexRu)):
		streamingService = model.Yandex
	case strings.Contains(host, string(model.SpotifyCom)):
		streamingService = model.Spotify
	case strings.Contains(host, string(model.DeezerLink)) || strings.Contains(host, string(model.DeezerLinkShort)):
		streamingService = model.Deezer
	default:
		fmt.Println("not supported service")
		return model.Unsupported, fmt.Errorf(streamingService.Error())
	}

	return streamingService, nil
}

func (p *ParsingService) GetMusicData(link *url.URL) (*model.LinkData, *model.ParsedMusic, error) {

	streamingService, err := p.identifyMusicServiceByLink(link)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка определения сервиса: %w", err)
	}

	parsingData, err := p.getParsingData(link, streamingService)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения данных о треке: %w", err)
	}

	// Получаем ссылки на трек
	links, err := p.getLinks(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения ссылок на трек: %w", err)
	}

	// Получаем полные данные о треке
	parsedMusic, err := p.getAllData(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения полных данных о треке: %w", err)
	}

	log.Printf("Получены данные о треке: %s - %s", parsingData.Title, parsingData.Artist)

	return links, parsedMusic, nil
}

func NewParsingService(fullParser IFullParsingMusicService, services ...IParsingMusicService) IParsingService {
	return &ParsingService{
		fullParser: fullParser,
		services:   services,
	}
}

func (p *ParsingService) getLinks(title, artist string) (*model.LinkData, error) {
	ld := &model.LinkData{Links: make(map[string]model.LinkResult)}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Для каждого музыкального сервиса запускаем горутину
	for _, svc := range p.services {
		wg.Add(1)
		go func(service IParsingMusicService) {
			defer wg.Done()
			// Получаем ссылку
			trackUrl, err := service.GetUrl(title, artist)

			// Определяем "название" сервиса по типу (для демонстрации)
			serviceName := fmt.Sprintf("%T", service)
			// Легкое логирование с юмором
			if err != nil {
				log.Printf("Ой, %s не смог найти ссылку для '%s - %s': %v", serviceName, title, artist, err)
			} else {
				log.Printf("Ура! %s нашел ссылку: %s", serviceName, trackUrl)
			}

			// Безопасно записываем результат в карту
			mu.Lock()
			ld.Links[serviceName] = model.LinkResult{
				URL:     trackUrl,
				Service: service.Service(),
				Error:   err,
			}
			mu.Unlock()
		}(svc)
	}
	// Ждем, пока все горутины завершат работу
	wg.Wait()

	// Если требуется, можно добавить дополнительную "фишку": например, проверку, что хоть один сервис успешно вернул ссылку.
	if len(ld.Links) == 0 {
		return nil, fmt.Errorf("ни один сервис не вернул ссылку для '%s - %s'", title, artist)
	}

	return ld, nil
}

func (p *ParsingService) getParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error) {
	var selectedService IParsingMusicService

	// Ищем нужный сервис среди зарегистрированных
	for _, svc := range p.services {
		if svc.Service() == streamingService {
			selectedService = svc
			break
		}
	}
	if selectedService == nil {
		return nil, fmt.Errorf("сервис %v не найден", streamingService)
	}

	trackUrl := link.Path
	title, artist, err := selectedService.GetData(trackUrl)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить данные от сервиса %v: %w", streamingService, err)
	}

	log.Printf("Сервис %v успешно обработал ссылку %s и вернул данные: %s - %s", streamingService, trackUrl, title, artist)

	return &model.ParsingData{
		Title:  title,
		Artist: artist,
	}, nil
}

func (p *ParsingService) getAllData(title, artist string) (*model.ParsedMusic, error) {

	parsedMusic, err := p.fullParser.GetAllData(title, artist)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения полных данных: %w", err)
	}

	return parsedMusic, nil
}
