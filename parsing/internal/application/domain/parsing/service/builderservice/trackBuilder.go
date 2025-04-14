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

type TrackBuilder struct {
	track model.Track

	MusicRepos   repository.MusicRepos
	LinkFormater modelutils.ILinkFormater
}

func NewTrackBuilder() *TrackBuilder {
	return &TrackBuilder{}
}

func (b *TrackBuilder) SetArtists(ctx context.Context, parsingArtists []model.ParsingArtist) error {

	for _, parsingArtist := range parsingArtists {
		artist, err := b.MusicRepos.ArtistRepo.GetArtistByName(ctx, parsingArtist.Name)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				newArtist := model.NewArtist(parsingArtist.Name, parsingArtist.Cover)

				if err := b.MusicRepos.ArtistRepo.Create(ctx, &newArtist); err != nil {
					return err
				}

				b.track.Artists = append(b.track.Artists, newArtist)
			} else {
				return err
			}

		} else {

			if artist != nil {
				b.track.Artists = append(b.track.Artists, *artist)
			}
		}
	}

	return nil
}

// before setting Album
func (b *TrackBuilder) SetGenres(ctx context.Context, parsingGenres []model.ParsingGenre) error {
	for _, parsingGenre := range parsingGenres {
		genre, err := b.MusicRepos.GenreRepo.GetGenreByName(ctx, parsingGenre.Name)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				newGenre := model.NewGenre(parsingGenre.Name)

				if err := b.MusicRepos.GenreRepo.Create(ctx, &newGenre); err != nil {
					return err
				}

				b.track.Album.Genres = append(b.track.Album.Genres, newGenre)
			} else {
				return err
			}

		} else {

			if genre != nil {
				b.track.Album.Genres = append(b.track.Album.Genres, *genre)
			}
		}
	}

	return nil
}

func (b *TrackBuilder) SetAlbum(ctx context.Context, albumName string, cover string, releaseData time.Time, recordType string) error {

	// Extract artist IDs from music.Artists
	artistIDs := make([]string, len(b.track.Artists))
	for i, artist := range b.track.Artists {
		artistIDs[i] = artist.ID
	}

	album, err := b.MusicRepos.AlbumRepo.GetAlbumByArtistsIDAndTitle(ctx, artistIDs, albumName)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			newAlbum := model.NewAlbum(albumName, cover, releaseData, recordType)
			if err := b.MusicRepos.AlbumRepo.Create(ctx, &newAlbum); err != nil {
				return err
			}

			b.track.Album = newAlbum
		} else {
			return err
		}
	} else {

		if album != nil {
			b.track.Album = *album
		}
	}

	// mb add relation Album -> Artist there

	return nil
}

func (b *TrackBuilder) SetTracks(ctx context.Context, parsingTracks []model.ParsingTrack) error {
	for _, parsingTrack := range parsingTracks {

		track, err := b.MusicRepos.TrackRepo.GetTrackByAlbumIDAndTrackName(ctx, b.track.AlbumID, parsingTrack.Title)

		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {

				newTrack := model.NewTrack(parsingTrack.Title, parsingTrack.DurationSec, parsingTrack.TrackPosition, parsingTrack.DiscNumber, parsingTrack.ReleaseDate)

				if err := b.MusicRepos.TrackRepo.Create(ctx, &newTrack); err != nil {
					return err
				}

				// set fields for track
				b.track.ID = newTrack.ID
				b.track.Title = newTrack.Title
				b.track.DurationSec = newTrack.DurationSec
				b.track.DiscNumber = newTrack.DiscNumber
				b.track.TrackPosition = newTrack.TrackPosition
				b.track.ReleaseDate = newTrack.ReleaseDate

			} else {
				return err
			}

		} else {
			if track != nil {
				// set fields for track
				b.track.ID = track.ID
				b.track.Title = track.Title
				b.track.DurationSec = track.DurationSec
				b.track.DiscNumber = track.DiscNumber
				b.track.TrackPosition = track.TrackPosition
				b.track.ReleaseDate = track.ReleaseDate
			}
		}
	}

	return nil
}

func (b *TrackBuilder) SetMusicLinks(ctx context.Context, links map[string]model.LinkResult) error {
	for _, link := range links {
		musicLink := model.NewMusicLink(link.Service, link.URL, b.track.ID, string(model.MusicRequestTrack))

		if err := b.MusicRepos.MusicLinkRepo.Create(ctx, &musicLink); err != nil {
			return err
		}

		formattedOutputLink, err := b.LinkFormater.FormatLinkFromDBToOutput(link.URL, link.Service)
		if err != nil {
			return err
		}

		// add formated link to output result
		musicLink.URL = *formattedOutputLink

		b.track.MusicLinks = append(b.track.MusicLinks, musicLink)
	}

	return nil
}

func (b *TrackBuilder) SetTrackPreview(ctx context.Context, musicPreviewURL string) error {

	newPreview := model.NewTrackPreview(b.track.ID, musicPreviewURL)
	if err := b.MusicRepos.TrackPreviewRepo.Create(ctx, &newPreview); err != nil {
		return err
	}

	b.track.TrackPreview = newPreview

	return nil
}

func (b *TrackBuilder) GetMusic() (*dto.MusicResult, error) {
	var track model.Track

	// format links
	formattedLinks, err := b.LinkFormater.FormatLinksFromDBToOutput(track.MusicLinks)
	if err != nil {
		return nil, err
	}

	track.MusicLinks = *formattedLinks

	// setting type
	track.Type = string(model.MusicRequestTrack)

	data, err := json.Marshal(track)
	if err != nil {
		return nil, err
	}

	music := &dto.MusicResult{
		Data: data,
	}

	return music, nil
}
