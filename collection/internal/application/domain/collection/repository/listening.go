package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type IListeningRepository interface {
	Create(ctx context.Context, listening *model.Listening) error
	WithTransaction(handler func() error) error
}

type ListeningRepository struct {
	db ports.DBPort
}

func (r *ListeningRepository) Create(ctx context.Context, listening *model.Listening) error {
	return r.db.Create(ctx, listening)
}

func (r *ListeningRepository) WithTransaction(handler func() error) error {
	return r.db.WithTransaction(handler)
}

func NewListeningRepository(db ports.DBPort) IListeningRepository {
	return &ListeningRepository{
		db: db,
	}
}
