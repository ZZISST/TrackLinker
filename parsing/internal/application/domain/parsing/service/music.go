package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gorm.io/gorm"
)

type IMusicService interface {
	GetTrackByID(ctx context.Context, id string) (*dto.MusicTrack, error)
	GetAlbumByID(ctx context.Context, id string) (*dto.MusicAlbum, error)
	FindMusic(link url.URL) (*dto.MusicResultDto, error)
	findTrack(ctx context.Context, link url.URL, mtype model.MusicRequestType) (*dto.MusicTrack, error)
	findAlbum(ctx context.Context, link url.URL) (*dto.MusicAlbum, error)
	upsertTrack(ctx context.Context, data *model.ParsedMusic, links *model.LinkData, mtype model.MusicRequestType) (*dto.MusicTrack, error)
}

type MusicService struct {
	TrackRepo        repository.ITrackRepository
	AlbumRepo        repository.IAlbumRepository
	ArtistRepo       repository.IArtistRepository
	MusicLinkRepo    repository.IMusicLinkRepository
	TrackPreviewRepo repository.ITrackPreviewRepository
	AlbumArtistRepo  repository.IAlbumArtistRepository
	ParsingSvc       IParsingService
	CloudSvc         ICloudService
	LinkFormater     modelutils.ILinkFormater
}

// GetAlbumByID implements IMusicService.
func (m MusicService) GetAlbumByID(ctx context.Context, id string) (*dto.MusicAlbum, error) {

	album, err := m.AlbumRepo.GetAlbumByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.MusicAlbum{
		AlbumID:    album.ID,
		Name:       album.Title,
		Artists:    album.Artists,
		CoverImage: album.CoverImage,
		ReleasedAt: album.ReleasedAt,
		CreatedAt:  album.CreatedAt,
		Links:      album.MusicLinks,
		Tracks:     album.Tracks,
	}, nil
}

// findAlbum implements IMusicService.
func (m MusicService) findAlbum(ctx context.Context, link url.URL) (*dto.MusicAlbum, error) {
	var music *dto.MusicAlbum

	album, err := m.AlbumRepo.GetAlbumByMusicLink(ctx, link.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Album: %+v, Error: %v", album, err)
	}

	return music, nil
}

// findTrack implements IMusicService.
func (m MusicService) findTrack(ctx context.Context, link url.URL, mtype model.MusicRequestType) (*dto.MusicTrack, error) {

	var music *dto.MusicTrack

	track, err := m.TrackRepo.GetTrackByMusicLink(context.Background(), link)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Track: %+v, Error: %v", track, err)

		linkData, fullData, error := m.ParsingSvc.GetMusicData(&link)
		if error != nil {
			return nil, fmt.Errorf("failed to get music data: %w", error)
		}

		newMusic, err := m.upsertTrack(context.Background(), fullData, linkData, mtype)
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

		for i := range track.MusicLinks {
			formattedOutputLink, err := m.LinkFormater.FormatLinkFromDBToOutput(track.MusicLinks[i].URL, track.MusicLinks[i].Service)
			if err != nil {
				return nil, fmt.Errorf("failed to format link from db to output: %w", err)
			}

			track.MusicLinks[i].URL = *formattedOutputLink
		}

		music = &dto.MusicTrack{
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

func (m MusicService) FindMusic(link url.URL) (*dto.MusicResultDto, error) {

	var music *dto.MusicResultDto

	musicType, err := m.LinkFormater.IdentifyMusicTypeByLink(&link)
	if err != nil {
		return nil, fmt.Errorf("failed to identify music type by link: %w", err)
	}

	switch *musicType {
	case model.MusicRequestTrack, model.MusicRequestTrackInAlbum:

		track, err := m.findTrack(context.Background(), link, *musicType)
		if err != nil {
			return nil, fmt.Errorf("failed to find track: %w", err)
		}
		data, err := json.Marshal(track)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal track: %w", err)
		}
		music = &dto.MusicResultDto{
			Type: string(model.MusicRequestTrack),
			Data: data,
		}

	case model.MusicRequestAlbum:

		album, err := m.findAlbum(context.Background(), link)
		if err != nil {
			return nil, fmt.Errorf("failed to find album: %w", err)
		}
		data, err := json.Marshal(album)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal album: %w", err)
		}
		music = &dto.MusicResultDto{
			Type: string(model.MusicRequestAlbum),
			Data: data,
		}
	default:
		return nil, fmt.Errorf("wrong link")
	}

	return music, nil
}

// GetTrackByID implements IMusicService.
func (m MusicService) GetTrackByID(ctx context.Context, id string) (*dto.MusicTrack, error) {

	track, err := m.TrackRepo.GetTrackByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.MusicTrack{
		TrackID:      track.ID,
		Album:        track.Album,
		Name:         track.Title,
		Artists:      track.Album.Artists,
		Links:        track.MusicLinks,
		TrackPreview: track.TrackPreview,
	}, nil
}

// upsertTrack implements IMusicService.
func (m MusicService) upsertTrack(ctx context.Context, data *model.ParsedMusic, links *model.LinkData, mtype model.MusicRequestType) (*dto.MusicTrack, error) {
	var music dto.MusicTrack

	// Define the transaction handler
	handler := func() error {

		log.Printf("Parsed Music: %+v", data)
		log.Printf("Link Data: %+v", links.Links)

		// Костыль потому что FullParser может быть не Deezer УЧЕСТЬ (можно impl GetUrl для Deezer
		deezerUrl, err := url.Parse(data.Link)
		if err != nil {
			return fmt.Errorf("failed to parse url: %w", err)
		}

		deezerMType, err := m.LinkFormater.IdentifyMusicTypeByLink(deezerUrl)
		if err != nil {
			return fmt.Errorf("failed to identify deezer type: %w", err)
		}

		formattedDeezer, err := m.LinkFormater.FormatLinkForDB(*deezerMType, deezerUrl)
		if err != nil {
			return fmt.Errorf("failed to format deezer to db: %w", err)
		}

		links.Links[data.Link] = model.LinkResult{
			URL:     *formattedDeezer,
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

		for _, artist := range music.Artists {
			if err := m.AlbumArtistRepo.AddRelation(ctx, music.Album.ID, artist.ID); err != nil {
				log.Printf("Failed to add relation between album %s and artist %s: %v", music.Album.ID, artist.ID, err)
				return fmt.Errorf("failed to add relation for artist %s: %w", artist.ID, err)
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

			musicLink := model.NewMusicLink(link.Service, link.URL, newTrack.ID, string(model.OwnerTypeTrack))

			if err := m.MusicLinkRepo.Create(ctx, &musicLink); err != nil {
				return fmt.Errorf("failed to upsert music link: %w", err)
			}

			formattedOutputLink, err := m.LinkFormater.FormatLinkFromDBToOutput(link.URL, link.Service)
			if err != nil {
				return fmt.Errorf("failed to format link from db to output: %w", err)
			}

			musicLink.URL = *formattedOutputLink

			music.Links = append(music.Links, musicLink)
		}

		return nil
	}

	// Execute the transaction
	err := m.TrackRepo.WithTransaction(handler)
	if err != nil {
		return nil, err
	}

	return &music, nil
}

func NewMusicService(trackRepo repository.ITrackRepository, artistRepo repository.IArtistRepository, albumRepo repository.IAlbumRepository, musicLinkRepo repository.IMusicLinkRepository, trackPreviewRepo repository.ITrackPreviewRepository, albumArtistRepo repository.IAlbumArtistRepository, parsingSvc IParsingService, cloudSvc ICloudService) IMusicService {
	return MusicService{
		TrackRepo:        trackRepo,
		AlbumRepo:        albumRepo,
		ArtistRepo:       artistRepo,
		MusicLinkRepo:    musicLinkRepo,
		TrackPreviewRepo: trackPreviewRepo,
		AlbumArtistRepo:  albumArtistRepo,
		ParsingSvc:       parsingSvc,
		CloudSvc:         cloudSvc,
		LinkFormater:     modelutils.NewLinkFormater(),
	}
}
