package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gorm.io/gorm"
)

type IMusicService interface {
	GetTrackByID(ctx context.Context, id string) (*model.Track, error)
	GetAlbumByID(ctx context.Context, id string) (*model.Album, error)
	GetAlbumTrackByID(ctx context.Context, albumId string, trackId string) (*model.Track, error)
	FindMusic(link url.URL) (*dto.MusicResult, error)
	findTrackByLink(ctx context.Context, link url.URL, mtype model.MusicRequestType) (*model.Track, error)
	findAlbumByLink(ctx context.Context, link url.URL) (*model.Album, error)
	findAlbumTrackByLink(ctx context.Context, link url.URL, mtype model.MusicRequestType) (*model.Track, error)
	upsertTrack(ctx context.Context, data *model.ParsingTrack, links *model.LinkData, mtype model.MusicRequestType) (*model.Track, error)
	// todo
	// upsertAlbum(ctx context.Context, data *model.ParsedMusic, links *model.LinkData, mtype model.MusicRequestType) (*model.Track, error)
	findTrackByData(ctx context.Context, name string, artists []string) (*model.Track, error)
}

type MusicService struct {
	TrackRepo        repository.ITrackRepository
	AlbumRepo        repository.IAlbumRepository
	ArtistRepo       repository.IArtistRepository
	MusicLinkRepo    repository.IMusicLinkRepository
	TrackPreviewRepo repository.ITrackPreviewRepository
	AlbumArtistRepo  repository.IAlbumArtistRepository
	GenreRepo        repository.IGenreRepository
	ParsingSvc       IParsingService
	CloudSvc         ICloudService
	LinkFormater     modelutils.ILinkFormater
}

func (m MusicService) GetTrackByID(ctx context.Context, id string) (*model.Track, error) {

	track, err := m.TrackRepo.GetTrackByID(ctx, id)
	if err != nil {
		return nil, err
	}

	for i := range track.MusicLinks {
		formattedOutputLink, err := m.LinkFormater.FormatLinkFromDBToOutput(track.MusicLinks[i].URL, track.MusicLinks[i].Service)
		if err != nil {
			return nil, fmt.Errorf("failed to format link from db to output: %w", err)
		}

		track.MusicLinks[i].URL = *formattedOutputLink
	}

	return track, nil
}

func (m MusicService) GetAlbumByID(ctx context.Context, id string) (*model.Album, error) {

	album, err := m.AlbumRepo.GetAlbumByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// format album links
	for i := range album.MusicLinks {
		formattedOutputLink, err := m.LinkFormater.FormatLinkFromDBToOutput(album.MusicLinks[i].URL, album.MusicLinks[i].Service)
		if err != nil {
			return nil, fmt.Errorf("failed to format link from db to output: %w", err)
		}

		album.MusicLinks[i].URL = *formattedOutputLink
	}

	// Format album's tracks links
	for i := range album.Tracks {
		track := &album.Tracks[i] // Get a pointer to the current track
		for j := range track.MusicLinks {
			formattedOutputLink, err := m.LinkFormater.FormatLinkFromDBToOutput(track.MusicLinks[j].URL, track.MusicLinks[j].Service)
			if err != nil {
				return nil, fmt.Errorf("failed to format link from db to output: %w", err)
			}

			// Perform formatting or processing on the music link
			track.MusicLinks[j].URL = *formattedOutputLink
		}
	}

	return album, nil
}

func (m MusicService) GetAlbumTrackByID(ctx context.Context, albumId string, trackId string) (*model.Track, error) {

	track, err := m.TrackRepo.GetTrackByAlbumAndTrackID(ctx, albumId, trackId)
	if err != nil {
		return nil, err
	}

	for i := range track.MusicLinks {
		formattedOutputLink, err := m.LinkFormater.FormatLinkFromDBToOutput(track.MusicLinks[i].URL, track.MusicLinks[i].Service)
		if err != nil {
			return nil, fmt.Errorf("failed to format link from db to output: %w", err)
		}

		track.MusicLinks[i].URL = *formattedOutputLink
	}

	return track, nil
}

func (m MusicService) FindMusic(link url.URL) (*dto.MusicResult, error) {

	var music *dto.MusicResult

	musicType, err := m.LinkFormater.IdentifyMusicTypeByLink(&link)
	if err != nil {
		return nil, fmt.Errorf("failed to identify music type by link: %w", err)
	}

	switch *musicType {
	case model.MusicRequestTrack:

		track, err := m.findTrackByLink(context.Background(), link, *musicType)
		if err != nil {
			return nil, fmt.Errorf("failed to find track: %w", err)
		}
		data, err := json.Marshal(track)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal track: %w", err)
		}
		music = &dto.MusicResult{
			Data: data,
		}

	case model.MusicRequestAlbum:

		album, err := m.findAlbumByLink(context.Background(), link)
		if err != nil {
			return nil, fmt.Errorf("failed to find album: %w", err)
		}
		data, err := json.Marshal(album)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal album: %w", err)
		}
		music = &dto.MusicResult{
			Data: data,
		}

	case model.MusicRequestTrackInAlbum:

		track, err := m.findAlbumTrackByLink(context.Background(), link, *musicType)
		if err != nil {
			return nil, fmt.Errorf("failed to find track in album: %w", err)
		}
		data, err := json.Marshal(track)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal track: %w", err)
		}
		music = &dto.MusicResult{
			Data: data,
		}

	default:
		return nil, fmt.Errorf("wrong link")
	}

	return music, nil
}

func (m MusicService) findTrackByLink(ctx context.Context, link url.URL, mtype model.MusicRequestType) (*model.Track, error) {

	var music *model.Track

	track, err := m.TrackRepo.GetTrackByMusicLink(context.Background(), link)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Track: %+v, Error: %v", track, err)

		linkData, fullData, error := m.ParsingSvc.GetTrackDataByLink(&link)
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
		music = track
	}

	return music, nil
}

func (m MusicService) findAlbumByLink(ctx context.Context, link url.URL) (*model.Album, error) {
	var music *model.Album

	album, err := m.AlbumRepo.GetAlbumByMusicLink(ctx, link.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {

		// make
		log.Printf("Album: %+v, Error: %v", album, err)
	}

	return music, nil
}

func (m MusicService) findAlbumTrackByLink(ctx context.Context, link url.URL, mtype model.MusicRequestType) (*model.Track, error) {
	var music *model.Track

	track, err := m.TrackRepo.GetTrackByMusicLink(context.Background(), link)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Track: %+v, Error: %v", track, err)

		linkData, fullData, error := m.ParsingSvc.GetTrackDataByLink(&link)
		if error != nil {
			return nil, fmt.Errorf("failed to get music data: %w", error)
		}

		// upsert album and album this track after we need in Background insert other..

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
		music = track
	}

	return music, nil
}

// In this method we know that THERE IS NO TRACK IN DB and maybe some Artists too! Uses by collection microservice
func (m MusicService) findTrackByData(ctx context.Context, name string, artists []string) (*model.Track, error) {
	// artists[0] полная хуйня надо менять на artists []string
	artistNames := make([]string, len(artists))

	// Extract artist names from music.Artists
	for i, artistName := range artists {
		artistNames[i] = artistName // Assuming each artist has a Name field
	}

	// Join the artist names into a single string, separated by commas
	artistStr := strings.Join(artistNames, ", ")

	linkData, fullData, err := m.ParsingSvc.GetTrackDataByData(&name, &artistStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get music data by data: %w", err)
	}

	newMusic, err := m.upsertTrack(context.Background(), fullData, linkData, model.MusicRequestTrack)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert track: %w", err)
	}

	return newMusic, nil
}

// upsertTrack implements IMusicService.
func (m MusicService) upsertTrack(ctx context.Context, data *model.ParsingTrack, links *model.LinkData, mtype model.MusicRequestType) (*model.Track, error) {
	var music model.Track

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

		// Костыль потому что FullParser может быть не Deezer УЧЕСТЬ (можно impl GetUrl для Deezer
		links.Links[data.Link] = model.LinkResult{
			URL:     *formattedDeezer,
			Service: model.Deezer,
		}

		// Несколько артистов
		for i := range data.Artists {
			artist, err := m.ArtistRepo.GetArtistByName(ctx, data.Artists[i].Name)

			if err != nil {

				if errors.Is(err, gorm.ErrRecordNotFound) {
					newArtist := model.NewArtist(data.Artists[i].Name, data.Artists[i].Cover)
					if err := m.ArtistRepo.Create(ctx, &newArtist); err != nil {
						return fmt.Errorf("failed to create artist: %w", err)
					}
					log.Printf("New Artist: %+v", newArtist)
					music.Artists = append(music.Artists, newArtist)
				}

				return fmt.Errorf("failed to get artist by name: %w", err)
			} else {
				if artist != nil {
					music.Artists = append(music.Artists, *artist)
					log.Printf(artist.Name)
				}
			}
		}

		// Check Genre
		for i := range data.Album.Genres {
			genre, err := m.GenreRepo.GetGenreByName(ctx, data.Album.Genres[i].Name)

			if err != nil {

				if errors.Is(err, gorm.ErrRecordNotFound) {
					newGenre := model.NewGenre(data.Album.Genres[i].Name)
					if err := m.GenreRepo.Create(ctx, &newGenre); err != nil {
						return fmt.Errorf("failed to create genre: %w", err)
					}
					log.Printf("New Genre: %+v", newGenre)
					music.Album.Genres = append(music.Album.Genres, newGenre)
				}

				return fmt.Errorf("failed to get genre by name: %w", err)
			} else {
				if genre != nil {
					music.Album.Genres = append(music.Album.Genres, *genre)
					log.Printf(genre.Name)
				}
			}
		}

		// Extract artist IDs from music.Artists
		artistIDs := make([]string, len(music.Artists))
		for i, artist := range music.Artists {
			artistIDs[i] = artist.ID // Assuming each artist has an ID field of type string
		}

		album, err := m.AlbumRepo.GetAlbumByArtistsIDAndTitle(ctx, artistIDs, data.Title)
		log.Printf("Album: %+v, Error: %v", album, err)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {

				newAlbum := model.NewAlbum(data.Title, data.Album.Cover, data.ReleaseDate, data.Album.RecordType)
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

		newTrack := model.NewTrack(data.Title, data.DurationSec, data.TrackPosition, data.DiscNumber, data.ReleaseDate)
		if err := m.TrackRepo.Create(ctx, &newTrack); err != nil {
			return fmt.Errorf("failed to create track: %w", err)
		}

		log.Printf("New Track: %+v", newTrack)

		// Upload to S3

		artistNames := make([]string, len(music.Artists))

		// Extract artist names from music.Artists
		for i, artist := range music.Artists {
			artistNames[i] = artist.Name // Assuming each artist has a Name field
		}

		// Join the artist names into a single string, separated by commas
		artistStr := strings.Join(artistNames, ", ")

		previewAudioUrl, err := m.CloudSvc.UploadTrackPreview(data.Preview, music.Title, artistStr)
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

			music.MusicLinks = append(music.MusicLinks, musicLink)
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

func NewMusicService(trackRepo repository.ITrackRepository, artistRepo repository.IArtistRepository, albumRepo repository.IAlbumRepository, musicLinkRepo repository.IMusicLinkRepository, trackPreviewRepo repository.ITrackPreviewRepository, albumArtistRepo repository.IAlbumArtistRepository, genreRepo repository.IGenreRepository, parsingSvc IParsingService, cloudSvc ICloudService) IMusicService {
	return MusicService{
		TrackRepo:        trackRepo,
		AlbumRepo:        albumRepo,
		ArtistRepo:       artistRepo,
		MusicLinkRepo:    musicLinkRepo,
		TrackPreviewRepo: trackPreviewRepo,
		AlbumArtistRepo:  albumArtistRepo,
		GenreRepo:        genreRepo,
		ParsingSvc:       parsingSvc,
		CloudSvc:         cloudSvc,
		LinkFormater:     modelutils.NewLinkFormater(),
	}
}
