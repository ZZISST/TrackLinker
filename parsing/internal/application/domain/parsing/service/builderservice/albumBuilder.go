package builderservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gorm.io/gorm"
)

type AlbumBuilder struct {
	album model.Album

	MusicRepos   repository.MusicRepos
	LinkFormater modelutils.ILinkFormater
}

func NewAlbumBuilder() *AlbumBuilder {
	return &AlbumBuilder{}
}

func (b *AlbumBuilder) SetArtists(ctx context.Context, parsingArtists []model.ParsingArtist) error {

	for _, parsingArtist := range parsingArtists {
		artist, err := b.MusicRepos.ArtistRepo.GetArtistByName(ctx, parsingArtist.Name)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				newArtist := model.NewArtist(parsingArtist.Name, parsingArtist.Cover)

				if err := b.MusicRepos.ArtistRepo.Create(ctx, &newArtist); err != nil {
					return fmt.Errorf("failed to create artist: %w", err)
				}

				b.album.Artists = append(b.album.Artists, newArtist)
			} else {
				return fmt.Errorf("failed to get artist by name: %w", err)
			}

		} else {

			if artist != nil {
				b.album.Artists = append(b.album.Artists, *artist)
			}
		}
	}

	return nil
}

func (b *AlbumBuilder) SetGenres(ctx context.Context, parsingGenres []model.ParsingGenre) error {
	for _, parsingGenre := range parsingGenres {
		genre, err := b.MusicRepos.GenreRepo.GetGenreByName(ctx, parsingGenre.Name)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				newGenre := model.NewGenre(parsingGenre.Name)

				if err := b.MusicRepos.GenreRepo.Create(ctx, &newGenre); err != nil {
					return fmt.Errorf("failed to create genre: %w", err)
				}

				b.album.Genres = append(b.album.Genres, newGenre)
			} else {
				return fmt.Errorf("failed to get genre by name: %w", err)
			}

		} else {

			if genre != nil {
				b.album.Genres = append(b.album.Genres, *genre)
			}
		}
	}

	return nil
}

func (b *AlbumBuilder) SetAlbum(ctx context.Context, albumName string, cover string, releaseData time.Time, recordType string, artists []model.Artist) error {

	// Extract artist IDs from music.Artists
	artistIDs := make([]string, len(b.album.Artists))
	for i, artist := range b.album.Artists {
		artistIDs[i] = artist.ID
	}

	album, err := b.MusicRepos.AlbumRepo.GetAlbumByArtistsIDAndTitle(ctx, artistIDs, albumName)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			newAlbum := model.NewAlbum(albumName, cover, releaseData, recordType)
			if err := b.MusicRepos.AlbumRepo.Create(ctx, &newAlbum); err != nil {
				return fmt.Errorf("failed to create album: %w", err)
			}

			// set fields for album
			b.album.ID = newAlbum.ID
			b.album.Title = newAlbum.Title
			b.album.CoverImage = newAlbum.CoverImage
			b.album.ReleaseDate = newAlbum.ReleaseDate
			b.album.RecordType = newAlbum.RecordType

		} else {
			return fmt.Errorf("failed to get album by name: %w", err)
		}
	} else {

		if album != nil {

			// set fields for album
			b.album.ID = album.ID
			b.album.Title = album.Title
			b.album.CoverImage = album.CoverImage
			b.album.ReleaseDate = album.ReleaseDate
			b.album.RecordType = album.RecordType

		}
	}

	// mb add relation Album -> Artist there

	return nil
}

func (b *AlbumBuilder) SetTracks(ctx context.Context, parsingTracks []model.ParsingTrack, albumID string) error {
	for _, parsingTrack := range parsingTracks {

		track, err := b.MusicRepos.TrackRepo.GetTrackByAlbumIDAndTrackName(ctx, albumID, parsingTrack.Title)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				newTrack := model.NewTrack(parsingTrack.Title, parsingTrack.DurationSec, parsingTrack.TrackPosition, parsingTrack.DiscNumber, parsingTrack.ReleaseDate)

				if err := b.MusicRepos.TrackRepo.Create(ctx, &newTrack); err != nil {
					return fmt.Errorf("failed to create track: %w", err)
				}

				b.album.Tracks = append(b.album.Tracks, newTrack)
			} else {
				return fmt.Errorf("failed to get track by name: %w", err)
			}

		} else {

			if track != nil {
				b.album.Tracks = append(b.album.Tracks, *track)
			}

		}
	}

	return nil
}

func (b *AlbumBuilder) SetMusicLinks(ctx context.Context, links map[string]model.LinkResult, ownerID string, ownerType model.OwnerType) error {
	for _, link := range links {
		musicLink := model.NewMusicLink(link.Service, link.URL, ownerID, string(ownerType))

		if err := b.MusicRepos.MusicLinkRepo.Create(ctx, &musicLink); err != nil {
			return fmt.Errorf("failed to upsert music link: %w", err)
		}

		formattedOutputLink, err := b.LinkFormater.FormatLinkFromDBToOutput(link.URL, link.Service)
		if err != nil {
			return fmt.Errorf("failed to format link from db to output: %w", err)
		}

		// add formated link to output result
		musicLink.URL = *formattedOutputLink

		b.album.MusicLinks = append(b.album.MusicLinks, musicLink)
	}

	return nil
}

func (b *AlbumBuilder) SetTrackPreview(context.Context, string, string, model.OwnerType) error {
	// no preview for tracks while requesting albums
	return nil
}

func (b *AlbumBuilder) GetMusic() (*dto.MusicResult, error) {
	var album model.Album

	// setting type
	album.Type = string(model.MusicRequestAlbum)

	data, err := json.Marshal(album)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal album: %w", err)
	}

	music := &dto.MusicResult{
		Data: data,
	}

	return music, nil
}
