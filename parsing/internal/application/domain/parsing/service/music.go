package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gorm.io/gorm"
)

type IMusicService interface {
	GetTrackByID(ctx context.Context, id string) (*dto.Music, error)
	FindTrack(link url.URL) (*dto.Music, error)
	upsertTrack(ctx context.Context, data *model.ParsedMusic, links *model.LinkData) (*dto.Music, error)
}

type MusicService struct {
	TrackRepo        repository.ITrackRepository
	AlbumRepo        repository.IAlbumRepository
	ArtistRepo       repository.IArtistRepository
	MusicLinkRepo    repository.IMusicLinkRepository
	TrackPreviewRepo repository.ITrackPreviewRepository
	ParsingSvc       IParsingService
	CloudSvc         ICloudService
}

// FindTrack implements IMusicService.
func (m MusicService) FindTrack(link url.URL) (*dto.Music, error) {

	var music *dto.Music

	// Retrieve the track by its music link
	track, err := m.TrackRepo.GetTrackByMusicLink(context.Background(), link.String())

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Track: %+v, Error: %v", track, err)

		linkData, fullData, error := m.ParsingSvc.GetMusicData(&link)
		if error != nil {
			return nil, fmt.Errorf("failed to get music data: %w", error)
		}

		newMusic, err := m.upsertTrack(context.Background(), fullData, linkData)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert track: %w", err)
		}

		return newMusic, nil

	} else if err != nil {
		return nil, fmt.Errorf("failed to find track by music link: %w", err)
	}

	// If the track is not found, return an error
	if track == nil {
		return nil, fmt.Errorf("track not found for link: %s", link.String())
	} else {
		music = &dto.Music{
			SongID:       track.ID,
			Album:        track.Album,
			Name:         track.Title,
			Artists:      track.Album.Artists,
			Links:        track.MusicLinks,
			TrackPreview: track.TrackPreview,
		}
	}

	return music, nil
}

// GetTrackByID implements IMusicService.
func (m MusicService) GetTrackByID(ctx context.Context, id string) (*dto.Music, error) {

	track, err := m.TrackRepo.GetTrackByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.Music{
		SongID:       track.ID,
		Album:        track.Album,
		Name:         track.Title,
		Artists:      track.Album.Artists,
		Links:        track.MusicLinks,
		TrackPreview: track.TrackPreview,
	}, nil
}

// upsertTrack implements IMusicService.
func (m MusicService) upsertTrack(ctx context.Context, data *model.ParsedMusic, links *model.LinkData) (*dto.Music, error) {
	var music dto.Music

	// Define the transaction handler
	handler := func() error {

		log.Printf("Parsed Music: %+v", data)

		artist, err := m.ArtistRepo.GetArtistByName(ctx, data.Artist.Name)
		log.Printf("Artist: %+v, Error: %v", artist, err)
		if err != nil {
			log.Printf("Error: %v", err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newArtist := model.NewArtist(data.Artist.Name, data.Artist.Link)
				if err := m.ArtistRepo.Create(ctx, &newArtist); err != nil {
					return fmt.Errorf("failed to create artist: %w", err)
				}
				log.Printf("New Artist: %+v", newArtist)
				music.Artists = append(music.Artists, newArtist)
			} else {
				return fmt.Errorf("failed to get artist by name: %w", err)
			}
		} else {

			if artist != nil {
				music.Artists = append(music.Artists, *artist)
				log.Printf(music.Artists[0].Name)
			}
		}

		album, err := m.AlbumRepo.GetAlbumByArtistIDAndTitle(ctx, music.Artists[0].ID, data.Title)
		log.Printf("Album: %+v, Error: %v", album, err)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {

				newAlbum := model.NewAlbum(data.Title, music.Artists[0].ID, data.Album.Cover, time.Now())
				if err := m.AlbumRepo.Create(ctx, &newAlbum); err != nil {
					return fmt.Errorf("failed to create album: %w", err)
				}
				music.Album = newAlbum
				log.Printf("New Album: %+v", newAlbum)
			} else {
				return fmt.Errorf("failed to get album by name: %w", err)
			}
		} else {
			if album != nil {
				music.Album = *album
			}
		}

		newTrack := model.NewTrack(data.Title, music.Album.ID)
		if err := m.TrackRepo.Create(ctx, &newTrack); err != nil {
			return fmt.Errorf("failed to create track: %w", err)
		}

		log.Printf("New Track: %+v", newTrack)

		music.Name = newTrack.Title
		music.SongID = newTrack.ID

		log.Printf("Music: %+v", music)

		previewAudioUrl, err := m.CloudSvc.UploadTrackPreview(data.Preview, music.Name, music.Artists[0].Name)
		if err != nil {
			return fmt.Errorf("failed to upload track preview: %w", err)
		}

		log.Printf("Preview Audio URL: %s", *previewAudioUrl)

		newPreview := model.NewTrackPreview(newTrack.ID, *previewAudioUrl)

		if err := m.TrackPreviewRepo.Create(ctx, &newPreview); err != nil {
			return fmt.Errorf("failed to create track preview: %w", err)
		}

		music.TrackPreview = newPreview

		log.Printf("Preview Audio URL: %s", music.TrackPreview.FileURL)

		// Upsert Music Links
		for _, link := range links.Links {
			musicLink := model.NewMusicLink(link.Service, link.URL, newTrack.ID)

			if err := m.MusicLinkRepo.Create(ctx, &musicLink); err != nil {
				return fmt.Errorf("failed to upsert music link: %w", err)
			}

			music.Links = append(music.Links, musicLink)
		}

		return nil
	}

	// Execute the transaction
	err := m.TrackRepo.WithTransaction(handler)
	if err != nil {
		return nil, err
	}

	// Return the DTO
	return &dto.Music{
		SongID:       music.SongID,
		Album:        music.Album,
		Name:         music.Name,
		Artists:      music.Artists,
		Links:        music.Links,
		TrackPreview: music.TrackPreview,
	}, nil
}

func NewMusicService(trackRepo repository.ITrackRepository, artistRepo repository.IArtistRepository, albumRepo repository.IAlbumRepository, musicLinkRepo repository.IMusicLinkRepository, trackPreviewRepo repository.ITrackPreviewRepository, parsingSvc IParsingService, cloudSvc ICloudService) IMusicService {
	return MusicService{
		TrackRepo:        trackRepo,
		AlbumRepo:        albumRepo,
		ArtistRepo:       artistRepo,
		MusicLinkRepo:    musicLinkRepo,
		TrackPreviewRepo: trackPreviewRepo,
		ParsingSvc:       parsingSvc,
		CloudSvc:         cloudSvc,
	}
}

// type MusicService struct {
// 	S3UploadingService cloud.S3UploadingService
// 	trackRepo          repositories.TrackRepositoryInterface
// }

// func (s *CocnreteParsingService) GetTrackByID(id string) (track *model.Track, err error) {
// 	trackID, err := uuid.Parse(id)
// 	if err != nil {
// 		return
// 	}

// 	track, err = s.trackRepo.GetTrackByID(trackID)
// 	if err != nil {
// 		return
// 	}

// 	return track, nil
// }

// func (s *CocnreteParsingService) GetTrackByLink(link string) (track *model.Track, err error) {

// 	parsedURL, err := url.Parse(link)
// 	if err != nil {
// 		return
// 	}

// 	musicService, err := s.identifyMusicServiceByUrl(parsedURL)
// 	if err != nil {
// 		return
// 	}

// 	track, err = s.getTrackByMusicURL(parsedURL, musicService)
// 	if err != nil {
// 		return
// 	}

// 	parsingData := s.getData(parsedURL, musicService)
// 	if parsingData.DataError != nil {
// 		return
// 	}

// 	songData, err := s.getFullData(parsingData.Title, parsingData.Artist)
// 	if err != nil {
// 		return
// 	}

// 	linkData := s.getLinks(parsingData.Title, parsingData.Artist)
// 	if linkData.AppleMusicError != nil {

// 	} else if linkData.SpotifyError != nil {

// 	} else if linkData.YandexMusicError != nil {

// 	} else if linkData.DeezerError != nil {

// 	}

// 	preview, err := s.uploadMp3ToCloud(songData.Link, songData.Title, songData.Artist.Name)
// 	if err != nil {
// 		return
// 	}

// 	return &model.Track{
// 		ID:              uuid.New(),
// 		Title:           songData.Title,
// 		Artist:          songData.Artist.Name,
// 		CoverURL:        songData.Album.Cover,
// 		PreviewAudioURL: *preview,
// 		DeezerURL:       linkData.DeezerURL,
// 		SpotifyURL:      linkData.SpotifyURL,
// 		AppleMusicURL:   linkData.AppleMusicURL,
// 		YandexMusicURL:  &linkData.YandexMusicURL,
// 	}, nil
// }

// func (s *CocnreteParsingService) uploadMp3ToCloud(mp3URL, title, artist string) (previewAudioUrl *string, err error) {
// 	previewAudioUrl, err = s.S3UploadingService.UploadMp3ToS3(mp3URL, title, artist)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return previewAudioUrl, nil
// }

// func (s *CocnreteParsingService) getFullData(title, artist string) (*Song, error) {
// 	song, err := s.DeezerService.GetAllData(title, artist)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return (*Song)(song), nil
// }

// func (s *CocnreteParsingService) identifyMusicServiceByUrl(link *url.URL) (musicService model.MusicService, err error) {
// 	var streamingService model.MusicService

// 	host := link.Host

// 	// identify music service
// 	switch {
// 	case strings.Contains(host, string(model.AppleCom)):
// 		streamingService = model.Apple
// 	case strings.Contains(host, string(model.YandexCom)) || strings.Contains(host, string(model.YandexRu)):
// 		streamingService = model.Yandex
// 	case strings.Contains(host, string(model.SpotifyCom)):
// 		streamingService = model.Spotify
// 	case strings.Contains(host, string(model.DeezerLink)) || strings.Contains(host, string(model.DeezerLinkShort)):
// 		streamingService = model.Deezer
// 	default:
// 		fmt.Println("not supported service")
// 		return model.Unsupported, fmt.Errorf(streamingService.Error())
// 	}

// 	return streamingService, nil
// }

// func (s *CocnreteParsingService) getTrackByMusicURL(link *url.URL, streamingService model.MusicService) (track *model.Track, err error) {

// 	existingTrack, err := s.trackRepo.GetTrackByURL(link.Host, streamingService)

// 	if !errors.Is(err, utils.NoTrackInDbError) {
// 		return nil, err
// 	}

// 	return existingTrack, nil
// }

// func (s *CocnreteParsingService) getData(link *url.URL, streamingService model.MusicService) ParsingData {
// 	var data ParsingData
// 	var err error
// 	switch streamingService {
// 	case model.Apple:
// 		data.Title, data.Artist, err = s.AppleService.GetData(link.Host)
// 		if err != nil {
// 			data.DataError = fmt.Errorf("ошибка apple music")
// 			return data
// 		}
// 		return data
// 	case model.Yandex:
// 		data.Title, data.Artist, err = s.YandexService.GetData(link.Host)
// 		if err != nil {
// 			data.DataError = fmt.Errorf("ошибка yandex music")
// 			return data
// 		}
// 		return data
// 	case model.Spotify:
// 		data.Title, data.Artist, err = s.SpotifyService.GetData(link.Host)
// 		if err != nil {
// 			data.DataError = fmt.Errorf("ошибка spotify")
// 			return data
// 		}
// 		return data
// 	case model.Deezer:
// 		data.Title, data.Artist, err = s.DeezerService.GetData(link.Host)
// 		if err != nil {
// 			data.DataError = fmt.Errorf("ошибка deezer")
// 			return data
// 		}
// 		return data
// 	default:
// 		return data
// 	}
// }

// func (s *CocnreteParsingService) getLinks(title, artist string) LinkData {
// 	var wg sync.WaitGroup
// 	links := LinkData{}
// 	mu := sync.Mutex{}

// 	wg.Add(4)

// 	go func() {
// 		defer wg.Done()
// 		deezer, err := s.DeezerService.GetUrl(title, artist)

// 		mu.Lock()
// 		links.DeezerURL = deezer
// 		links.DeezerError = err
// 		mu.Unlock()
// 	}()

// 	go func() {
// 		defer wg.Done()
// 		spotify, err := s.SpotifyService.GetUrl(title, artist)

// 		mu.Lock()
// 		links.SpotifyURL = spotify
// 		links.SpotifyError = err
// 		mu.Unlock()
// 	}()

// 	go func() {
// 		defer wg.Done()
// 		apple, err := s.AppleService.GetUrl(title, artist)

// 		mu.Lock()
// 		links.AppleMusicURL = apple
// 		links.AppleMusicError = err
// 		mu.Unlock()
// 	}()

// 	go func() {
// 		defer wg.Done()
// 		yandex, err := s.YandexService.GetUrl(title, artist)
// 		mu.Lock()
// 		links.YandexMusicURL = yandex
// 		links.YandexMusicError = err
// 		mu.Unlock()
// 	}()

// 	wg.Wait()
// 	fmt.Println(links)
// 	return links
// }

// func NewParsingService(
// 	apple service.MusicServiceInterface,
// 	yandex service.MusicServiceInterface,
// 	spotify service.MusicServiceInterface,
// 	deezer service.MusicServiceFullParsingInterface,
// 	uploadingService cloud.S3UploadingService,
// 	trackRepo repositories.TrackRepositoryInterface,
// ) ParsingServiceInterface {
// 	return &CocnreteParsingService{
// 		AppleService:       apple,
// 		YandexService:      yandex,
// 		SpotifyService:     spotify,
// 		DeezerService:      deezer,
// 		S3UploadingService: uploadingService,
// 		trackRepo:          trackRepo,
// 	}
// }
