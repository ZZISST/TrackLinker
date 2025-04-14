package service

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service/musicservice"
)

type IParsingService interface {
	getTrackLinks(title string, artist string) (*model.LinkData, error)
	getAlbumLinks(title string, artist string) (*model.LinkData, error)
	getAllTrackData(title, artist string) (*model.ParsingTrack, error)
	getAllAlbumData(title, artist string) (*model.ParsingAlbum, error)
	getTrackParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error)
	getAlbumParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error)
	GetTrackDataByLink(link *url.URL) (*model.LinkData, *model.ParsingTrack, error)
	GetTrackDataByData(title *string, artist *string) (*model.LinkData, *model.ParsingTrack, error)
	GetAlbumDataByLink(link *url.URL) (*model.LinkData, *model.ParsingAlbum, error)
	GetAlbumDataByData(title *string, artist *string) (*model.LinkData, *model.ParsingAlbum, error)
}

type ParsingService struct {
	services     []musicservice.IParsingMusicService
	fullParser   musicservice.IFullParsingMusicService
	linkFormater modelutils.ILinkFormater
}

func NewParsingService(fullParser musicservice.IFullParsingMusicService, services ...musicservice.IParsingMusicService) IParsingService {
	return &ParsingService{
		fullParser:   fullParser,
		services:     services,
		linkFormater: modelutils.NewLinkFormater(),
	}
}

func (p *ParsingService) GetTrackDataByData(title *string, artist *string) (*model.LinkData, *model.ParsingTrack, error) {

	if title == nil || artist == nil {
		return nil, nil, fmt.Errorf("неправильный ввод артиста и имени трэка")
	}

	// Получаем ссылки на трек
	links, err := p.getTrackLinks(*title, *artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения ссылок на трек: %w", err)
	}

	// Получаем полные данные о треке
	parsedMusic, err := p.getAllTrackData(*title, *artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения полных данных о треке: %w", err)
	}

	log.Printf("Получены данные о треке: %s - %s", *title, *artist)

	return links, parsedMusic, nil
}

func (p *ParsingService) GetTrackDataByLink(link *url.URL) (*model.LinkData, *model.ParsingTrack, error) {

	log.Printf("Получение данных о треке по ссылке: %s", link)
	streamingService, err := p.linkFormater.IdentifyMusicServiceByLink(link)

	if err != nil {
		return nil, nil, fmt.Errorf("ошибка определения сервиса: %w", err)
	}
	log.Printf("Определен сервис: %v", streamingService)

	parsingData, err := p.getTrackParsingData(link, streamingService)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения данных о треке: %w", err)
	}
	log.Printf("Получены данные о треке: %s - %s", parsingData.Title, parsingData.Artist)

	// Получаем ссылки на трек
	links, err := p.getTrackLinks(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения ссылок на трек: %w", err)
	}

	// Получаем полные данные о треке
	parsedMusic, err := p.getAllTrackData(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения полных данных о треке: %w", err)
	}

	log.Printf("Получены данные о треке: %s - %s", parsingData.Title, parsingData.Artist)

	return links, parsedMusic, nil
}

func (p *ParsingService) GetAlbumDataByLink(link *url.URL) (*model.LinkData, *model.ParsingAlbum, error) {
	log.Printf("Получение данных об альбоме по ссылке: %s", link)
	streamingService, err := p.linkFormater.IdentifyMusicServiceByLink(link)

	if err != nil {
		return nil, nil, fmt.Errorf("ошибка определения сервиса: %w", err)
	}
	log.Printf("Определен сервис: %v", streamingService)

	parsingData, err := p.getAlbumParsingData(link, streamingService)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения данных об альбоме: %w", err)
	}
	log.Printf("Получены данные об альбоме: %s - %s", parsingData.Title, parsingData.Artist)

	// Получаем ссылки на альбом
	links, err := p.getAlbumLinks(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения ссылок на трек: %w", err)
	}

	// Получаем полные данные об альбоме
	parsedMusic, err := p.getAllAlbumData(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения полных данных об альбоме: %w", err)
	}

	log.Printf("Получены данные о треке: %s - %s", parsingData.Title, parsingData.Artist)

	return links, parsedMusic, nil
}

func (p *ParsingService) GetAlbumDataByData(title *string, artist *string) (*model.LinkData, *model.ParsingAlbum, error) {
	if title == nil || artist == nil {
		return nil, nil, fmt.Errorf("неправильный ввод артиста и имени альбома")
	}

	// Получаем ссылки на трек
	links, err := p.getAlbumLinks(*title, *artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения ссылок на альбом: %w", err)
	}

	// Получаем полные данные о треке
	parsedMusic, err := p.getAllAlbumData(*title, *artist)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка получения полных данных об альбоме: %w", err)
	}

	log.Printf("Получены данные об альбоме: %s - %s", *title, *artist)

	return links, parsedMusic, nil
}

func (p *ParsingService) getTrackLinks(title string, artist string) (*model.LinkData, error) {
	ld := &model.LinkData{Links: make(map[string]model.LinkResult)}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Для каждого музыкального сервиса запускаем горутину
	for _, svc := range p.services {
		wg.Add(1)
		go func(service musicservice.IParsingMusicService) {
			defer wg.Done()
			// Получаем ссылку

			urlStr, err := service.GetTrackUrl(title, artist)
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

func (p *ParsingService) getTrackParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error) {
	var selectedService musicservice.IParsingMusicService

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
	title, artist, err := selectedService.GetTrackData(trackUrl)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить данные от сервиса %v: %w", streamingService, err)
	}

	log.Printf("Сервис %v успешно обработал ссылку %s и вернул данные: %s - %s", streamingService, trackUrl, title, artist)

	return &model.ParsingData{
		Title:  title,
		Artist: artist,
	}, nil
}

func (p *ParsingService) getAlbumParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error) {
	var selectedService musicservice.IParsingMusicService

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
	title, artist, err := selectedService.GetAlbumData(trackUrl)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить данные от сервиса %v: %w", streamingService, err)
	}

	log.Printf("Сервис %v успешно обработал ссылку %s и вернул данные: %s - %s", streamingService, trackUrl, title, artist)

	return &model.ParsingData{
		Title:  title,
		Artist: artist,
	}, nil
}

func (p *ParsingService) getAllTrackData(title, artist string) (*model.ParsingTrack, error) {

	parsedMusic, err := p.fullParser.GetAllTrackData(title, artist)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения полных данных: %w", err)
	}

	return parsedMusic, nil
}

func (p *ParsingService) getAlbumLinks(title string, artist string) (*model.LinkData, error) {
	ld := &model.LinkData{Links: make(map[string]model.LinkResult)}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Для каждого музыкального сервиса запускаем горутину
	for _, svc := range p.services {
		wg.Add(1)
		go func(service musicservice.IParsingMusicService) {
			defer wg.Done()
			// Получаем ссылку

			urlStr, err := service.GetAlbumUrl(title, artist)
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

func (p *ParsingService) getAllAlbumData(title, artist string) (*model.ParsingAlbum, error) {
	parsedMusic, err := p.fullParser.GetAllAlbumData(title, artist)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения полных данных: %w", err)
	}

	return parsedMusic, nil
}
