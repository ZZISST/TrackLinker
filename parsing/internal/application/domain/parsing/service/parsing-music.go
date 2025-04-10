package service

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
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
	getLinks(title string, artist string) (*model.LinkData, error)
	getAllData(title, artist string) (*model.ParsedMusic, error)
	getParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error)
	GetMusicDataByLink(link *url.URL) (*model.LinkData, *model.ParsedMusic, error)
	GetMusicDataByData(title *string, artist *string) (*model.LinkData, *model.ParsedMusic, error)
}

type ParsingService struct {
	services     []IParsingMusicService
	fullParser   IFullParsingMusicService
	linkFormater modelutils.ILinkFormater
}

func NewParsingService(fullParser IFullParsingMusicService, services ...IParsingMusicService) IParsingService {
	return &ParsingService{
		fullParser:   fullParser,
		services:     services,
		linkFormater: modelutils.NewLinkFormater(),
	}
}

func (p *ParsingService) GetMusicDataByData(title *string, artist *string) (*model.LinkData, *model.ParsedMusic, error) {

	if title == nil || artist == nil {
		return nil, nil, fmt.Errorf("неправильный ввод артиста и имени трэка")
	}

	// Получаем ссылки на трек
	links, err := p.getLinks(*title, *artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения ссылок на трек: %w", err)
	}

	// Получаем полные данные о треке
	parsedMusic, err := p.getAllData(*title, *artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения полных данных о треке: %w", err)
	}

	log.Printf("Получены данные о треке: %s - %s", *title, *artist)

	return links, parsedMusic, nil
}

func (p *ParsingService) GetMusicDataByLink(link *url.URL) (*model.LinkData, *model.ParsedMusic, error) {

	log.Printf("Получение данных о треке по ссылке: %s", link)
	streamingService, err := p.linkFormater.IdentifyMusicServiceByLink(link)

	if err != nil {
		return nil, nil, fmt.Errorf("ошибка определения сервиса: %w", err)
	}
	log.Printf("Определен сервис: %v", streamingService)

	parsingData, err := p.getParsingData(link, streamingService)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения данных о треке: %w", err)
	}
	log.Printf("Получены данные о треке: %s - %s", parsingData.Title, parsingData.Artist)

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

func (p *ParsingService) getLinks(title string, artist string) (*model.LinkData, error) {
	ld := &model.LinkData{Links: make(map[string]model.LinkResult)}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Для каждого музыкального сервиса запускаем горутину
	for _, svc := range p.services {
		wg.Add(1)
		go func(service IParsingMusicService) {
			defer wg.Done()
			// Получаем ссылку

			urlStr, err := service.GetUrl(title, artist)
			if err != nil {
				log.Printf("Не удалось получить ссылку для '%s - %s' от сервиса %v: %v", title, artist, service.Service(), err)
				return
			}

			trackUrl, err := url.Parse(urlStr)
			if err != nil {
				log.Printf("Неверная ссылка для '%s - %s' от сервиса %v: %v", title, artist, service.Service(), err)
				return
			}

			mtype, err := p.linkFormater.IdentifyMusicTypeByLink(trackUrl)
			if err != nil {
				log.Printf("Music Type не определен")
				return
			}

			formattedLink, err := p.linkFormater.FormatLinkForDB(*mtype, trackUrl)
			if err != nil {
				log.Printf("Неверная ссылка для '%s - %s'", title, artist)
				return
			}

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
				URL:     *formattedLink,
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
