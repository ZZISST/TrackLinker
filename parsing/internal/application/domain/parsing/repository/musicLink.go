package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type IMusicLinkRepository interface {
	Create(ctx context.Context, musicLink *model.MusicLink) error
}

type MusicLinkRepository struct {
	db ports.DBPort
}

func NewMusicLinkRepository(db ports.DBPort) IMusicLinkRepository {
	return &MusicLinkRepository{
		db: db,
	}
}

func (r *MusicLinkRepository) Create(ctx context.Context, musicLink *model.MusicLink) error {
	return r.db.Create(ctx, musicLink)
}
