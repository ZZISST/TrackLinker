package builderservice

import (
	"context"
	"encoding/json"
	"errors"
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
					return err
				}

				b.album.Artists = append(b.album.Artists, newArtist)
			} else {
				return err
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
					return err
				}

				b.album.Genres = append(b.album.Genres, newGenre)
			} else {
				return err
			}

		} else {

			if genre != nil {
				b.album.Genres = append(b.album.Genres, *genre)
			}
		}
	}

	return nil
}

func (b *AlbumBuilder) SetAlbum(ctx context.Context, albumName string, cover string, releaseData time.Time, recordType string) error {

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
				return err
			}

			// set fields for album
			b.album.ID = newAlbum.ID
			b.album.Title = newAlbum.Title
			b.album.CoverImage = newAlbum.CoverImage
			b.album.ReleaseDate = newAlbum.ReleaseDate
			b.album.RecordType = newAlbum.RecordType

		} else {
			return err
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

func (b *AlbumBuilder) SetTracks(ctx context.Context, parsingTracks []model.ParsingTrack) error {

	for _, parsingTrack := range parsingTracks {

		track, err := b.MusicRepos.TrackRepo.GetTrackByAlbumIDAndTrackName(ctx, b.album.ID, parsingTrack.Title)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				newTrack := model.NewTrack(parsingTrack.Title, parsingTrack.DurationSec, parsingTrack.TrackPosition, parsingTrack.DiscNumber, parsingTrack.ReleaseDate)

				if err := b.MusicRepos.TrackRepo.Create(ctx, &newTrack); err != nil {
					return err
				}

				b.album.Tracks = append(b.album.Tracks, newTrack)
			} else {
				return err
			}

		} else {

			if track != nil {
				b.album.Tracks = append(b.album.Tracks, *track)
			}

		}
	}

	return nil
}

func (b *AlbumBuilder) SetMusicLinks(ctx context.Context, links map[string]model.LinkResult) error {
	for _, link := range b.album.MusicLinks {
		musicLink := model.NewMusicLink(link.Service, link.URL, b.album.ID, string(model.MusicRequestAlbum))

		if err := b.MusicRepos.MusicLinkRepo.Create(ctx, &musicLink); err != nil {
			return err
		}

		formattedOutputLink, err := b.LinkFormater.FormatLinkFromDBToOutput(link.URL, link.Service)
		if err != nil {
			return err
		}

		// add formated link to output result
		musicLink.URL = *formattedOutputLink

		b.album.MusicLinks = append(b.album.MusicLinks, musicLink)
	}

	return nil
}

func (b *AlbumBuilder) SetTrackPreview(ctx context.Context, musicPreviewURL string) error {
	// add track previews for album's tracks
	for _, track := range b.album.Tracks {
		newPreview := model.NewTrackPreview(track.ID, musicPreviewURL)
		if err := b.MusicRepos.TrackPreviewRepo.Create(ctx, &newPreview); err != nil {
			return err
		}
	}
	return nil
}

func (b *AlbumBuilder) GetMusic() (*dto.MusicResult, error) {
	var album model.Album

	// setting type
	album.Type = string(model.MusicRequestAlbum)

	data, err := json.Marshal(album)
	if err != nil {
		return nil, err
	}

	music := &dto.MusicResult{
		Data: data,
	}

	return music, nil
}
