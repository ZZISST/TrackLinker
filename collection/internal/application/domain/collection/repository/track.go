package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model/modelMusic"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type ITrackRepository interface {
	WithTransaction(handler func() error) error
	GetTrackByNameAndArtistsID(ctx context.Context, name string, artistID []string) (*modelMusic.Track, error)
}

type TrackRepository struct {
	db ports.DBPort
}

func (r *TrackRepository) WithTransaction(handler func() error) error {
	return r.db.WithTransaction(handler)
}

func (r *TrackRepository) GetTrackByNameAndArtistsID(ctx context.Context, name string, artistID []string) (*modelMusic.Track, error) {
	var track modelMusic.Track
	opts := []ports.FindOption{
		ports.WithPreload([]string{
			"Album",
			"Album.Artists",
			"TrackPreview",
			"MusicLinks",
		}),
		ports.WithQuery(ports.NewQuery("tracks.title = ? AND artists.id IN ?", name, artistID)),
	}

	if err := r.db.FindOne(ctx, &track, opts...); err != nil {
		return nil, err
	}
	return &track, nil
}

func NewTrackRepository(db ports.DBPort) ITrackRepository {
	return &TrackRepository{
		db: db,
	}
}
