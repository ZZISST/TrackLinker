package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/modelutils"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
	"gorm.io/gorm"
)

type ITrackRepository interface {
	Create(ctx context.Context, track *model.Track) error
	GetTrackByID(ctx context.Context, trackID string) (*model.Track, error)
	GetTrackByAlbumAndTrackID(ctx context.Context, albumID string, trackID string) (*model.Track, error)
	GetTrackByMusicLink(ctx context.Context, musicLink url.URL) (*model.Track, error)
	GetTrackInAlbumByMusicLink(ctx context.Context, musicLink url.URL) (*model.Track, error)
	WithTransaction(handler func() error) error
}

type TrackRepository struct {
	db           ports.DBPort
	linkFormater modelutils.ILinkFormater
}

// GetTrackByMusicLink implements ITrackRepository.
func (r *TrackRepository) GetTrackByMusicLink(ctx context.Context, musicLink url.URL) (*model.Track, error) {
	var track model.Track

	dbFormatLink, err := r.linkFormater.FormatLinkForDB(model.MusicRequestTrack, &musicLink)
	if err != nil {
		return nil, err
	}

	log.Printf("DB format link: %s", *dbFormatLink)

	opts := []ports.FindOption{
		ports.WithJoin("JOIN music_links ON music_links.owner_id = tracks.id"),
		ports.WithPreload([]string{
			"Album",
			"Album.Artists",
			"TrackPreview",
			"MusicLinks",
		}),
		ports.WithQuery(ports.NewQuery("music_links.url = ?", *dbFormatLink)), // Add owner_type condition
	}

	if err := r.db.FindOne(ctx, &track, opts...); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Track not found for music link: %s", musicLink)
			return nil, err // Track not found
		}
		return nil, err
	}

	return &track, nil
}

func (r *TrackRepository) GetTrackInAlbumByMusicLink(ctx context.Context, musicLink url.URL) (*model.Track, error) {
	var track model.Track

	dbFormatLink, err := r.linkFormater.FormatLinkForDB(model.MusicRequestTrackInAlbum, &musicLink)
	if err != nil {
		return nil, err
	}

	log.Printf("DB format link: %s", *dbFormatLink)

	opts := []ports.FindOption{
		ports.WithJoin("JOIN music_links ON music_links.owner_id = tracks.id"),
		ports.WithPreload([]string{
			"Album",
			"Album.Artists",
			"TrackPreview",
			"MusicLinks",
		}),
		ports.WithQuery(ports.NewQuery("music_links.url = ?", *dbFormatLink)), // Add owner_type condition
	}

	if err := r.db.FindOne(ctx, &track, opts...); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Track not found for music link: %s", musicLink.Host)
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
		ports.WithPreload([]string{
			"Album",
			"Album.Artists",
			"TrackPreview",
			"MusicLinks",
		}))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Трек не найден
		}
		return nil, err
	}

	return &track, nil
}

func (r *TrackRepository) GetTrackByAlbumAndTrackID(ctx context.Context, albumID string, trackID string) (*model.Track, error) {
	var track model.Track

	// Query the database for the track with the given albumID and trackID
	err := r.db.FindOne(ctx, &track,
		ports.WithQuery(ports.NewQuery("album_id = ? AND id = ?", albumID, trackID)),
		ports.WithPreload([]string{
			"Album",
			"Album.Artists",
			"TrackPreview",
			"MusicLinks",
		}))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("track not found for albumID: %s and trackID: %s", albumID, trackID)
		}
		return nil, fmt.Errorf("failed to fetch track: %w", err)
	}

	return &track, nil
}

func (r *TrackRepository) Create(ctx context.Context, track *model.Track) error {
	return r.db.Create(ctx, track)
}

func NewTrackRepository(db ports.DBPort) ITrackRepository {
	return &TrackRepository{
		db:           db,
		linkFormater: modelutils.NewLinkFormater(),
	}
}
