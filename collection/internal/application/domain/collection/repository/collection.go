package repository

import (
	"context"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type ICollectionRepository interface {
	GetCollectionByUserID(ctx context.Context, userID string) (*model.Collection, error)
	AppendAddedContent(ctx context.Context, content model.AddedContent)
}

type CollectionRepository struct {
	db ports.DBPort
}

func (r *CollectionRepository) GetCollectionByUserID(ctx context.Context, userID string) (*model.Collection, error) {
	var collection model.Collection

	opts := []ports.FindOption{
		ports.WithQuery(ports.NewQuery("user_id = ?", userID)),
	}

	if err := r.db.FindOne(ctx, &collection, opts...); err != nil {
		return nil, err
	}

	return &collection, nil
}

func (r *CollectionRepository) AppendAddedContent(ctx context.Context, content model.AddedContent) {

}

func NewCollectionRepository(db ports.DBPort) ICollectionRepository {
	return &CollectionRepository{
		db: db,
	}
}
