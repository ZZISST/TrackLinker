package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
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
	formatInputLink(link url.URL) (*string, error)
	identifyMusicServiceByLink(link *url.URL) (model.MusicService, error)
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

// identifyMusicServiceByLink implements IMusicService.
func (m MusicService) identifyMusicServiceByLink(link *url.URL) (model.MusicService, error) {
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
		return model.Unsupported, errors.New(streamingService.Error())
	}

	return streamingService, nil
}

// formatInputLink implements IMusicService.
func (m MusicService) formatInputLink(link url.URL) (*string, error) {

	var formattedInputLink string

	musicService, err := m.identifyMusicServiceByLink(&link)
	if err != nil {
		return nil, nil
	}

	pathParts := strings.Split(link.Path, "/")

	switch musicService {
	case model.Apple:
		// Apple Music
		// Example: https://music.apple.com/us/song/the-beautiful-people/1440747194
		region := pathParts[1]
		trackName := pathParts[len(pathParts)-2]
		trackId := pathParts[len(pathParts)-1]

		formattedInputLink = region + "/" + trackName + "/" + trackId

	case model.Yandex:
		// Yandex Music
		// Example: https://music.yandex.ru/track/100000

		if len(pathParts) < 5 {
			if pathParts[1] == "album" {
				albumID := pathParts[len(pathParts)-1]
				formattedInputLink = "album/" + albumID
			} else {
				trackID := pathParts[len(pathParts)-1]
				formattedInputLink = "track/" + trackID
			}
		} else {
			albumId := pathParts[len(pathParts)-2]
			trackID := pathParts[len(pathParts)-1]
			formattedInputLink = "album/" + albumId + "/" + trackID
		}

	case model.Spotify:
		// Spotify
		// Example: https://open.spotify.com/track/3ZFTkvIE7kyPt6Nu3PEa7V
		if pathParts[1] == "track" {
			trackID := pathParts[len(pathParts)-1]
			formattedInputLink = "track/" + trackID
		} else {
			albumID := pathParts[len(pathParts)-1]
			formattedInputLink = "album/" + albumID
		}
	case model.Deezer:
		// Deezer
		// Example: https://deezer.com/track/123456789
		trackID := pathParts[len(pathParts)-1]
		formattedInputLink = "track/" + trackID

	default:
		return nil, fmt.Errorf("unsupported music service")
	}

	return &formattedInputLink, nil
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
			TrackID:      track.ID,
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
		TrackID:      track.ID,
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

		// Костыль потому что FullParser может быть не Deezer УЧЕСТЬ (можно impl GetUrl для Deezer)
		links.Links[data.Link] = model.LinkResult{
			URL:     data.Link,
			Service: model.Deezer,
		}

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
		music.TrackID = newTrack.ID

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
		TrackID:      music.TrackID,
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
