package repository

import (
	"context"
	"errors"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
	"gorm.io/gorm"
)

type ITrackRepository interface {
	Create(ctx context.Context, track *model.Track) error
	GetTrackByID(ctx context.Context, trackID string) (*model.Track, error)
	GetTrackByMusicLink(ctx context.Context, musicLink string) (*model.Track, error)
	WithTransaction(handler func() error) error
}

type TrackRepository struct {
	db ports.DBPort
}

// GetTrackByMusicLink implements ITrackRepository.
func (r *TrackRepository) GetTrackByMusicLink(ctx context.Context, musicLink string) (*model.Track, error) {
	var track model.Track
	opts := []ports.FindOption{
		ports.WithJoin("JOIN music_links ON music_links.track_id = tracks.id"),
		ports.WithQuery(ports.NewQuery("music_links.url = ?", musicLink)),
		ports.WithPreload([]string{"Album", "Album.Artists", "MusicLinks", "TrackPreview"}), // Preload related data
	}

	if err := r.db.FindOne(ctx, &track, opts...); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err // Track not found
		}
		return nil, err
	}

	return &track, nil
}

func (r *TrackRepository) WithTransaction(handler func() error) error {
	return r.db.WithTransaction(handler)
}

// GetMusicByTrackID implements ITrackRepository.
func (r *TrackRepository) GetTrackByID(ctx context.Context, trackID string) (*model.Track, error) {
	var track model.Track

	err := r.db.FindOne(ctx, &track,
		ports.WithQuery(ports.NewQuery("id = ?", trackID)),
		ports.WithPreload([]string{"Album", "Album.Artists", "MusicLinks"}), // Загружаем связанные данные
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Трек не найден
		}
		return nil, err
	}

	return &track, nil
}

func NewTrackRepository(db ports.DBPort) ITrackRepository {
	return &TrackRepository{
		db: db,
	}
}

func (r *TrackRepository) Create(ctx context.Context, track *model.Track) error {
	return r.db.Create(ctx, track)
}
