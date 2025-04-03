package repository

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type IAlbumArtistRepository interface {
	AddRelation(ctx context.Context, albumID, artistID string) error
}

type AlbumArtistRepository struct {
	db ports.DBPort
}

func (r *AlbumArtistRepository) AddRelation(ctx context.Context, albumID, artistID string) error {
	// Check if the relation already exists
	var count int64
	query := "SELECT COUNT(*) FROM album_artists WHERE album_id = ? AND artist_id = ?"
	if err := r.db.GetDB().Raw(query, albumID, artistID).Scan(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing relation: %w", err)
	}

	// If the relation already exists, return without inserting
	if count > 0 {
		log.Printf("Relation between album %s and artist %s already exists", albumID, artistID)
		return nil
	}

	// Insert the new relation
	query = "INSERT INTO album_artists (album_id, artist_id) VALUES (?, ?)"
	if err := r.db.GetDB().Exec(query, albumID, artistID).Error; err != nil {
		return fmt.Errorf("failed to insert relation: %w", err)
	}

	log.Printf("Successfully added relation between album %s and artist %s", albumID, artistID)
	return nil
}

func NewAlbumArtistRepository(db ports.DBPort) IAlbumArtistRepository {
	return &AlbumArtistRepository{
		db: db,
	}
}
