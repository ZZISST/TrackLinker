package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type ITrackPreviewRepository interface {
	Create(ctx context.Context, preview *model.TrackPreview) error
}

type TrackPreviewRepository struct {
	db ports.DBPort
}

func NewTrackPreviewRepository(db ports.DBPort) ITrackPreviewRepository {
	return &TrackPreviewRepository{
		db: db,
	}
}

func (r *TrackPreviewRepository) Create(ctx context.Context, preview *model.TrackPreview) error {
	return r.db.Create(ctx, preview)
}
