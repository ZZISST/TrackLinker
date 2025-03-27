package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type IArtistRepository interface {
	Create(ctx context.Context, artist *model.Artist) error
	GetArtistByName(ctx context.Context, artistName string) (*model.Artist, error)
}

type ArtistRepository struct {
	db ports.DBPort
}

// GetArtistByName implements IArtistRepository.
func (r *ArtistRepository) GetArtistByName(ctx context.Context, artistName string) (*model.Artist, error) {
	var artist model.Artist

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

func (r *ArtistRepository) Create(ctx context.Context, artist *model.Artist) error {
	return r.db.Create(ctx, artist)
}
