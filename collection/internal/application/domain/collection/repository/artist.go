package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model/modelMusic"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type IArtistRepository interface {
	GetArtistByName(ctx context.Context, artistName string) (*modelMusic.Artist, error)
}

type ArtistRepository struct {
	db ports.DBPort
}

// GetArtistByName implements IArtistRepository.
func (r *ArtistRepository) GetArtistByName(ctx context.Context, artistName string) (*modelMusic.Artist, error) {
	var artist modelMusic.Artist

	opts := []ports.FindOption{
		ports.WithQuery(ports.NewQuery("name = ?", artistName)),
	}

	if err := r.db.FindOne(ctx, &artist, opts...); err != nil {
		return nil, err
	}

	return &artist, nil
}

func NewArtistRepository(db ports.DBPort) IArtistRepository {
	return &ArtistRepository{
		db: db,
	}
}
