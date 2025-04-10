package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type IAddedContentRepository interface {
	Create(ctx context.Context, listening *model.AddedContent) error
	GetByCollectionIDAndTrackID(ctx context.Context, collectionID, trackID string) (*model.AddedContent, error)
}

type AddedContentRepository struct {
	db ports.DBPort
}

func (r *AddedContentRepository) Create(ctx context.Context, addedContent *model.AddedContent) error {
	return r.db.Create(ctx, addedContent)
}

func (r *AddedContentRepository) GetByCollectionIDAndTrackID(ctx context.Context, collectionID, trackID string) (*model.AddedContent, error) {
	var addedContent model.AddedContent

	opts := []ports.FindOption{
		ports.WithQuery(ports.NewQuery("collection_id = ? AND track_id = ?", collectionID, trackID)),
	}

	if err := r.db.FindOne(ctx, &addedContent, opts...); err != nil {
		return nil, err
	}

	return &addedContent, nil
}

func NewAddedContentRepository(db ports.DBPort) IAddedContentRepository {
	return &AddedContentRepository{
		db: db,
	}
}
