package service

import (
	"fmt"
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
	GetAlbumTrackDataByLink(link *url.URL) (*model.LinkData, *model.ParsingTrack, error)
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
		return nil, nil, fmt.Errorf("Wrong input artist and album name")
	}

	// Получаем ссылки на трек
	links, err := p.getTrackLinks(*title, *artist)
	if err != nil {
		return nil, nil, err
	}

	// Получаем полные данные о треке
	parsedMusic, err := p.getAllTrackData(*title, *artist)
	if err != nil {
		return nil, nil, err
	}

	return links, parsedMusic, nil
}

func (p *ParsingService) GetTrackDataByLink(link *url.URL) (*model.LinkData, *model.ParsingTrack, error) {

	streamingService, err := p.linkFormater.IdentifyMusicServiceByLink(link)

	if err != nil {
		return nil, nil, err
	}

	parsingData, err := p.getTrackParsingData(link, streamingService)
	if err != nil {
		return nil, nil, err
	}

	links, err := p.getTrackLinks(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, err
	}

	parsedMusic, err := p.getAllTrackData(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, err
	}

	return links, parsedMusic, nil
}

func (p *ParsingService) GetAlbumDataByLink(link *url.URL) (*model.LinkData, *model.ParsingAlbum, error) {

	streamingService, err := p.linkFormater.IdentifyMusicServiceByLink(link)
	if err != nil {
		return nil, nil, err
	}

	parsingData, err := p.getAlbumParsingData(link, streamingService)
	if err != nil {
		return nil, nil, err
	}

	links, err := p.getAlbumLinks(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, err
	}

	parsedMusic, err := p.getAllAlbumData(parsingData.Title, parsingData.Artist)
	if err != nil {
		return nil, nil, err
	}

	return links, parsedMusic, nil
}

func (p *ParsingService) GetAlbumTrackDataByLink(link *url.URL) (*model.LinkData, *model.ParsingTrack, error) {
	return nil, nil, nil
}

func (p *ParsingService) GetAlbumDataByData(title *string, artist *string) (*model.LinkData, *model.ParsingAlbum, error) {

	if title == nil || artist == nil {
		return nil, nil, fmt.Errorf("Wrong input artist and album name")
	}

	links, err := p.getAlbumLinks(*title, *artist)
	if err != nil {
		return nil, nil, err
	}

	parsedMusic, err := p.getAllAlbumData(*title, *artist)
	if err != nil {
		return nil, nil, err
	}

	return links, parsedMusic, nil
}

func (p *ParsingService) getTrackLinks(title string, artist string) (*model.LinkData, error) {

	ld := &model.LinkData{Links: make(map[string]model.LinkResult)}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, svc := range p.services {
		wg.Add(1)
		go func(service musicservice.IParsingMusicService) {
			defer wg.Done()

			urlStr, err := service.GetTrackUrl(title, artist)
			if err != nil {
				return
			}

			trackUrl, err := url.Parse(urlStr)
			if err != nil {
				return
			}

			mtype, err := p.linkFormater.IdentifyMusicTypeByLink(trackUrl)
			if err != nil {
				return
			}

			formattedLink, err := p.linkFormater.FormatLinkForDB(*mtype, trackUrl)
			if err != nil {
				return
			}

			serviceName := string(service.Service())

			mu.Lock()
			ld.Links[serviceName] = model.LinkResult{
				URL:     *formattedLink,
				Service: service.Service(),
				Error:   err,
			}
			mu.Unlock()
		}(svc)
	}
	wg.Wait()

	return ld, nil
}

func (p *ParsingService) getTrackParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error) {
	var selectedService musicservice.IParsingMusicService

	for _, svc := range p.services {
		if svc.Service() == streamingService {
			selectedService = svc
			break
		}
	}
	if selectedService == nil {
		return nil, fmt.Errorf("Service %v not found", streamingService)
	}

	trackUrl := link.Path
	title, artist, err := selectedService.GetTrackData(trackUrl)
	if err != nil {
		return nil, err
	}

	return &model.ParsingData{
		Title:  title,
		Artist: artist,
	}, nil
}

func (p *ParsingService) getAlbumParsingData(link *url.URL, streamingService model.MusicService) (*model.ParsingData, error) {
	var selectedService musicservice.IParsingMusicService

	for _, svc := range p.services {
		if svc.Service() == streamingService {
			selectedService = svc
			break
		}
	}
	if selectedService == nil {
		return nil, fmt.Errorf("Service %v not found", streamingService)
	}

	trackUrl := link.Path
	title, artist, err := selectedService.GetAlbumData(trackUrl)
	if err != nil {
		return nil, err
	}

	return &model.ParsingData{
		Title:  title,
		Artist: artist,
	}, nil
}

func (p *ParsingService) getAllTrackData(title, artist string) (*model.ParsingTrack, error) {

	parsedMusic, err := p.fullParser.GetAllTrackData(title, artist)
	if err != nil {
		return nil, err
	}

	return parsedMusic, nil
}

func (p *ParsingService) getAlbumLinks(title string, artist string) (*model.LinkData, error) {
	ld := &model.LinkData{Links: make(map[string]model.LinkResult)}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, svc := range p.services {
		wg.Add(1)
		go func(service musicservice.IParsingMusicService) {
			defer wg.Done()

			urlStr, err := service.GetAlbumUrl(title, artist)
			if err != nil {
				return
			}

			trackUrl, err := url.Parse(urlStr)
			if err != nil {
				return
			}

			mtype, err := p.linkFormater.IdentifyMusicTypeByLink(trackUrl)
			if err != nil {
				return
			}

			formattedLink, err := p.linkFormater.FormatLinkForDB(*mtype, trackUrl)
			if err != nil {
				return
			}

			serviceName := string(service.Service())

			mu.Lock()
			ld.Links[serviceName] = model.LinkResult{
				URL:     *formattedLink,
				Service: service.Service(),
				Error:   err,
			}
			mu.Unlock()
		}(svc)
	}
	wg.Wait()

	return ld, nil
}

func (p *ParsingService) getAllAlbumData(title, artist string) (*model.ParsingAlbum, error) {
	parsedMusic, err := p.fullParser.GetAllAlbumData(title, artist)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения полных данных: %w", err)
	}

	return parsedMusic, nil
}
