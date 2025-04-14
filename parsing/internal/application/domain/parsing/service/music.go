package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service/builderservice"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service/cloudservice"
	"gorm.io/gorm"
)

/*
IMusicService defines the interface for interacting with music-related data.

This interface includes methods for retrieving, finding, and upserting tracks and albums, as well as handling music links.

Author: Vladimir Stabrovskiy @godnesty
*/
type IMusicService interface {
	GetTrackByID(ctx context.Context, id string) (*model.Track, error)
	GetAlbumByID(ctx context.Context, id string) (*model.Album, error)
	GetAlbumTrackByID(ctx context.Context, albumId string, trackId string) (*model.Track, error)
	FindMusic(link url.URL) (*dto.MusicResult, error)
	findMusicByLink(ctx context.Context, link url.URL, musicType model.MusicRequestType) (*dto.MusicResult, error)
	findTrackByLink(ctx context.Context, link url.URL) (*dto.MusicResult, error)
	findAlbumByLink(ctx context.Context, link url.URL) (*dto.MusicResult, error)
	findAlbumTrackByLink(ctx context.Context, link url.URL) (*dto.MusicResult, error)
	FindTrackByData(ctx context.Context, name string, artists []string) (*model.Track, error)
}

/*
MusicService implements the IMusicService interface.

This struct provides methods for managing tracks, albums, and related metadata.
It interacts with various repositories and services to handle operations such as retrieving, finding,
and upserting music data.

Author: Vladimir Stabrovskiy @godnesty
*/
type MusicService struct {
	InputLink string

	MusicRepos   repository.MusicRepos
	ParsingSvc   IParsingService
	CloudSvc     cloudservice.ICloudService
	LinkFormater modelutils.ILinkFormater
}

/*
GetTrackByID retrieves a track by its ID.

Parameters:
  - ctx (context.Context): The context for the request.
  - id (string): The ID of the track.

Returns:
  - *model.Track: The track with the specified ID.
  - error: An error if the operation fails.Author: Vladimir Stabrovskiy @godnesty

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) GetTrackByID(ctx context.Context, id string) (*model.Track, error) {

	track, err := m.MusicRepos.TrackRepo.GetTrackByID(ctx, id)
	if err != nil {
		return nil, err
	}

	formattedLinks, err := m.LinkFormater.FormatLinksFromDBToOutput(track.MusicLinks)
	if err != nil {
		return nil, err
	}

	track.MusicLinks = *formattedLinks

	return track, nil
}

/*
GetAlbumByID retrieves an album by its ID.

Parameters:
  - ctx (context.Context): The context for the request.
  - id (string): The ID of the album.

Returns:
  - *model.Album: The album with the specified ID.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) GetAlbumByID(ctx context.Context, id string) (*model.Album, error) {

	album, err := m.MusicRepos.AlbumRepo.GetAlbumByID(ctx, id)
	if err != nil {
		return nil, err
	}

	formattedLinks, err := m.LinkFormater.FormatLinksFromDBToOutput(album.MusicLinks)
	if err != nil {
		return nil, err
	}

	album.MusicLinks = *formattedLinks

	// Format album's tracks links
	for _, track := range album.Tracks {

		formattedLinks, err = m.LinkFormater.FormatLinksFromDBToOutput(track.MusicLinks)
		if err != nil {
			return nil, err
		}

		track.MusicLinks = *formattedLinks

	}

	return album, nil
}

/*
GetAlbumTrackByID retrieves a track from an album by the album ID and track ID.

Parameters:
  - ctx (context.Context): The context for the request.
  - albumId (string): The ID of the album.
  - trackId (string): The ID of the track within the album.

Returns:
  - *model.Track: The track with the specified album and track IDs.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) GetAlbumTrackByID(ctx context.Context, albumId string, trackId string) (*model.Track, error) {

	track, err := m.MusicRepos.TrackRepo.GetTrackByAlbumAndTrackID(ctx, albumId, trackId)
	if err != nil {
		return nil, err
	}

	formattedLinks, err := m.LinkFormater.FormatLinksFromDBToOutput(track.MusicLinks)
	if err != nil {
		return nil, err
	}

	track.MusicLinks = *formattedLinks

	return track, nil
}

/*
FindMusic finds music metadata based on a given link.

Parameters:
  - link (url.URL): The URL of the music to find.

Returns:
  - *dto.MusicResult: The result containing music metadata.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) FindMusic(link url.URL) (*dto.MusicResult, error) {

	musicType, err := m.LinkFormater.IdentifyMusicTypeByLink(&link)
	if err != nil {
		return nil, err
	}

	music, err := m.findMusicByLink(context.Background(), link, *musicType)
	if err != nil {
		return nil, err
	}

	return music, nil
}

func (m MusicService) findMusicByLink(ctx context.Context, link url.URL, musicType model.MusicRequestType) (*dto.MusicResult, error) {

	switch musicType {

	case model.MusicRequestTrack:
		track, err := m.findTrackByLink(ctx, link)
		if err != nil {
			return nil, err
		}
		return track, nil

	case model.MusicRequestAlbum:
		album, err := m.findAlbumByLink(ctx, link)
		if err != nil {
			return nil, err
		}
		return album, nil

	case model.MusicRequestTrackInAlbum:
		track, err := m.findAlbumTrackByLink(ctx, link)
		if err != nil {
			return nil, err
		}
		return track, nil

	default:
		return nil, fmt.Errorf("unsupported music type: %s", musicType)
	}
}

/*
findTrackByLink finds a track by its link and request type.

Parameters:
  - ctx (context.Context): The context for the request.
  - link (url.URL): The URL of the track.
  - mtype (model.MusicRequestType): The type of music request.

Returns:
  - *model.Track: The track found by the link.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) findTrackByLink(ctx context.Context, link url.URL) (*dto.MusicResult, error) {
	var music *dto.MusicResult

	// Check track
	track, err := m.MusicRepos.TrackRepo.GetTrackByMusicLink(context.Background(), link)
	if errors.Is(err, gorm.ErrRecordNotFound) {

		// Get Data
		linkData, fullData, err := m.ParsingSvc.GetTrackDataByLink(&link)
		if err != nil {
			return nil, err
		}

		// Get TrackBuilder
		builder, err := builderservice.GetBuilder(model.MusicRequestTrack)
		if err != nil {
			return nil, err
		}

		// Convert fullData to a slice of model.ParsingTrack
		var parsingTrack []model.ParsingTrack
		if fullData != nil {
			parsingTrack = append(parsingTrack, *fullData)
		}

		// Upload preview to S3
		artistNames := make([]string, len(fullData.Artists))
		for i, artist := range fullData.Artists {
			artistNames[i] = artist.Name
		}
		artistStr := strings.Join(artistNames, ", ")
		previewAudioUrl, err := m.CloudSvc.UploadTrackPreview(fullData.Preview, fullData.Title, artistStr)
		if err != nil {
			return nil, err
		}

		// Insert Track
		builder.SetArtists(ctx, fullData.Artists)
		builder.SetGenres(ctx, fullData.Album.Genres)
		builder.SetAlbum(ctx, fullData.Album.Title, fullData.Album.Cover, fullData.Album.ReleaseDate, fullData.Album.RecordType)
		builder.SetTracks(ctx, parsingTrack)
		builder.SetTrackPreview(ctx, *previewAudioUrl)
		builder.SetMusicLinks(ctx, linkData.Links)

		music, err = builder.GetMusic()
		if err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	}

	if track != nil {
		formattedLinks, err := m.LinkFormater.FormatLinksFromDBToOutput(track.MusicLinks)
		if err != nil {
			return nil, err
		}

		track.MusicLinks = *formattedLinks

		data, err := json.Marshal(track)
		if err != nil {
			return nil, err
		}

		music = &dto.MusicResult{
			Data: data,
		}
	}

	return music, nil
}

/*
findAlbumByLink finds an album by its link and request type.

Parameters:
  - ctx (context.Context): The context for the request.
  - link (url.URL): The URL of the album.
  - mtype (model.MusicRequestType): The type of music request.

Returns:
  - *model.Album: The album found by the link.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) findAlbumByLink(ctx context.Context, link url.URL) (*dto.MusicResult, error) {
	var music *dto.MusicResult

	// Check album
	album, err := m.MusicRepos.AlbumRepo.GetAlbumByMusicLink(context.Background(), link)
	if errors.Is(err, gorm.ErrRecordNotFound) {

		// Get Data
		linkData, fullData, err := m.ParsingSvc.GetAlbumDataByLink(&link)
		if err != nil {
			return nil, err
		}

		// Get TrackBuilder
		builder, err := builderservice.GetBuilder(model.MusicRequestAlbum)
		if err != nil {
			return nil, err
		}

		// Insert Album
		builder.SetArtists(ctx, fullData.Artists)
		builder.SetGenres(ctx, fullData.Genres)
		builder.SetAlbum(ctx, fullData.Title, fullData.Cover, fullData.ReleaseDate, fullData.RecordType)
		builder.SetTracks(ctx, fullData.Tracks)
		builder.SetMusicLinks(ctx, linkData.Links)

		music, err = builder.GetMusic()
		if err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	}

	if album != nil {
		formattedLinks, err := m.LinkFormater.FormatLinksFromDBToOutput(album.MusicLinks)
		if err != nil {
			return nil, err
		}

		album.MusicLinks = *formattedLinks

		data, err := json.Marshal(album)
		if err != nil {
			return nil, err
		}

		music = &dto.MusicResult{
			Data: data,
		}
	}

	return music, nil
}

/*
findAlbumTrackByLink finds a track within an album by its link and request type.

Parameters:
  - ctx (context.Context): The context for the request.
  - link (url.URL): The URL of the track within the album.
  - mtype (model.MusicRequestType): The type of music request.

Returns:
  - *model.Track: The track found by the link.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) findAlbumTrackByLink(ctx context.Context, link url.URL) (*dto.MusicResult, error) {
	var music *dto.MusicResult

	// Check track

	track, err := m.MusicRepos.TrackRepo.GetTrackByMusicLink(context.Background(), link)
	if errors.Is(err, gorm.ErrRecordNotFound) {

		// Get Data
		linkData, fullData, err := m.ParsingSvc.GetTrackDataByLink(&link)
		if err != nil {
			return nil, err
		}

		// Get TrackBuilder
		builder, err := builderservice.GetBuilder(model.MusicRequestTrackInAlbum)
		if err != nil {
			return nil, err
		}

		// Convert fullData to a slice of model.ParsingTrack
		var parsingTrack []model.ParsingTrack
		if fullData != nil {
			parsingTrack = append(parsingTrack, *fullData)
		}

		// Upload preview to S3
		artistNames := make([]string, len(fullData.Artists))
		for i, artist := range fullData.Artists {
			artistNames[i] = artist.Name
		}
		artistStr := strings.Join(artistNames, ", ")
		previewAudioUrl, err := m.CloudSvc.UploadTrackPreview(fullData.Preview, fullData.Title, artistStr)
		if err != nil {
			return nil, err
		}

		// Insert Track
		builder.SetArtists(ctx, fullData.Artists)
		builder.SetGenres(ctx, fullData.Album.Genres)
		builder.SetAlbum(ctx, fullData.Album.Title, fullData.Album.Cover, fullData.Album.ReleaseDate, fullData.Album.RecordType)
		builder.SetTracks(ctx, parsingTrack)
		builder.SetTrackPreview(ctx, *previewAudioUrl)
		builder.SetMusicLinks(ctx, linkData.Links)

		music, err = builder.GetMusic()
		if err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	}

	if track != nil {
		formattedLinks, err := m.LinkFormater.FormatLinksFromDBToOutput(track.MusicLinks)
		if err != nil {
			return nil, err
		}

		track.MusicLinks = *formattedLinks

		data, err := json.Marshal(track)
		if err != nil {
			return nil, err
		}

		music = &dto.MusicResult{
			Data: data,
		}
	}

	return music, nil
}

/*
findTrackByData finds a track by its name and associated artists.

Parameters:
  - ctx (context.Context): The context for the request.
  - name (string): The name of the track.
  - artists ([]string): A list of artist names associated with the track.

Returns:
  - *model.Track: The track found by the name and artists.
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (m MusicService) FindTrackByData(ctx context.Context, name string, artists []string) (*model.Track, error) {

	artistNames := make([]string, len(artists))
	for i, artistName := range artists {
		artistNames[i] = artistName
	}
	artistStr := strings.Join(artistNames, ", ")

	linkData, fullData, err := m.ParsingSvc.GetTrackDataByData(&name, &artistStr)
	if err != nil {
		return nil, err
	}

	// Get TrackBuilder
	builder, err := builderservice.GetBuilder(model.MusicRequestTrack)
	if err != nil {
		return nil, err
	}

	// Convert fullData to a slice of model.ParsingTrack
	var parsingTrack []model.ParsingTrack
	if fullData != nil {
		parsingTrack = append(parsingTrack, *fullData)
	}

	// Upload preview to S3
	previewAudioUrl, err := m.CloudSvc.UploadTrackPreview(fullData.Preview, fullData.Title, artistStr)
	if err != nil {
		return nil, err
	}

	// Insert Track
	builder.SetArtists(ctx, fullData.Artists)
	builder.SetGenres(ctx, fullData.Album.Genres)
	builder.SetAlbum(ctx, fullData.Album.Title, fullData.Album.Cover, fullData.Album.ReleaseDate, fullData.Album.RecordType)
	builder.SetTracks(ctx, parsingTrack)
	builder.SetTrackPreview(ctx, *previewAudioUrl)
	builder.SetMusicLinks(ctx, linkData.Links)

	newTrack, err := builder.GetMusic()
	if err != nil {
		return nil, err
	}

	var track model.Track

	if err := json.Unmarshal(newTrack.Data, &track); err != nil {
		return nil, err
	}

	return &track, nil
}

/*
NewMusicService creates and returns a new instance of MusicService.

Parameters:
  - musicRepos
  - parsingSvc (IParsingService): The service for parsing music data.
  - cloudSvc (ICloudService): The service for cloud operations.

Returns:
  - IMusicService: A new instance of MusicService.

Author: Vladimir Stabrovskiy @godnesty
*/
func NewMusicService(musicRepos repository.MusicRepos, parsingSvc IParsingService, cloudSvc cloudservice.ICloudService) IMusicService {
	return MusicService{
		MusicRepos:   musicRepos,
		ParsingSvc:   parsingSvc,
		CloudSvc:     cloudSvc,
		LinkFormater: modelutils.NewLinkFormater(),
	}
}
