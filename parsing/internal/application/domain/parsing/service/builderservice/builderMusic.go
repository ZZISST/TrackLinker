package builderservice

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

type IBuilderMusic interface {
	SetArtists(ctx context.Context, parsingArtists []model.ParsingArtist) error
	SetGenres(ctx context.Context, parsingGenres []model.ParsingGenre) error
	SetAlbum(ctx context.Context, albumName string, cover string, releaseData time.Time, recordType string) error
	SetTracks(ctx context.Context, parsingTracks []model.ParsingTrack) error
	SetTrackPreview(ctx context.Context, musicPreviewURL string) error
	SetMusicLinks(ctx context.Context, links map[string]model.LinkResult) error
	GetMusic() (*dto.MusicResult, error)
}

func GetBuilder(musicType model.MusicRequestType) (IBuilderMusic, error) {
	switch musicType {
	case model.MusicRequestTrack:
		return NewTrackBuilder(), nil

	case model.MusicRequestTrackInAlbum:
		return NewTrackInAlbumBuilder(), nil

	case model.MusicRequestAlbum:
		return NewAlbumBuilder(), nil

	default:
		return nil, fmt.Errorf("wrong music type")
	}
}
