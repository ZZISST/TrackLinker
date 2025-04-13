package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type IGenreRepository interface {
	Create(ctx context.Context, genre *model.Genre) error
	GetGenreByName(ctx context.Context, name string) (*model.Genre, error)
}

type GenreRepository struct {
	db ports.DBPort
}

func (r *GenreRepository) Create(ctx context.Context, genre *model.Genre) error {
	return r.db.Create(ctx, genre)
}

func (r *GenreRepository) GetGenreByName(ctx context.Context, name string) (*model.Genre, error) {
	var genre model.Genre

	opts := []ports.FindOption{
		ports.WithQuery(ports.NewQuery("name = ?", name)),
	}

	if err := r.db.FindOne(ctx, &genre, opts...); err != nil {
		return nil, err
	}

	return &genre, nil
}

func NewGenreRepository(db ports.DBPort) IGenreRepository {
	return &GenreRepository{
		db: db,
	}
}
