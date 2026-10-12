package repository

import (
	"context"
	"fmt"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model/modelMusic"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type ITrackRepository interface {
	WithTransaction(handler func() error) error
	GetTrackByNameAndArtistsID(ctx context.Context, name string, artistIDs []string) (*modelMusic.Track, error)
}

type TrackRepository struct {
	db ports.DBPort
}

func (r *TrackRepository) WithTransaction(handler func() error) error {
	return r.db.WithTransaction(handler)
}

func (r *TrackRepository) GetTrackByNameAndArtistsID(ctx context.Context, name string, artistIDs []string) (*modelMusic.Track, error) {
	var tracks []modelMusic.Track

	opts := []ports.FindOption{
		ports.WithPreload([]string{
			"Album",
			"Album.Artists", // Ensure artists are preloaded from the album
			"TrackPreview",
			"MusicLinks",
		}),
		ports.WithQuery(ports.NewQuery("tracks.title = ?", name)), // Query by track title only
	}

	// Find the track
	if err := r.db.Find(ctx, &tracks, opts...); err != nil {
		return nil, err
	}

	// Iterate through the tracks and check if any match the artist IDs
	for _, track := range tracks {
		for _, artist := range track.Album.Artists {
			for _, artistID := range artistIDs {
				if artist.ID == artistID {
					return &track, nil // Return the first matching track
				}
			}
		}
	}

	return nil, fmt.Errorf("no matching artist found in album for track: %s", name)
}

func NewTrackRepository(db ports.DBPort) ITrackRepository {
	return &TrackRepository{
		db: db,
	}
}
