package repository

import (
	"context"
	"log"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type IAlbumRepository interface {
	Create(ctx context.Context, album *model.Album) error
	GetAlbumByArtistIDAndTitle(ctx context.Context, artistID, title string) (*model.Album, error)
	GetAlbumByMusicLink(ctx context.Context, musicLink string) (*model.Album, error)
	GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error)
}

type AlbumRepository struct {
	db ports.DBPort
}

// GetAlbumByID implements IAlbumRepository.
func (r *AlbumRepository) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	var album model.Album

	if err := r.db.FindOne(ctx, &album, ports.WithQuery(ports.NewQuery("id = ?", albumID))); err != nil {
		return nil, err
	}

	return &album, nil
}

// GetAlbumByMusicLink implements IAlbumRepository.
func (r *AlbumRepository) GetAlbumByMusicLink(ctx context.Context, musicLink string) (*model.Album, error) {
	var album model.Album

	opts := []ports.FindOption{
		ports.WithJoin("JOIN music_links ON music_links.owner_id = albums.id"),
		ports.WithQuery(ports.NewQuery("music_links.url ILIKE ?", musicLink)), // Use ILIKE for case-insensitive comparison
		ports.WithPreload([]string{"Artists", "MusicLinks", "Tracks"}),        // Preload related data
	}

	if err := r.db.FindOne(ctx, &album, opts...); err != nil {
		return nil, err
	}

	return &album, nil
}

// GetAlbumByArtistIDAndTitle implements IAlbumRepository.
func (r *AlbumRepository) GetAlbumByArtistIDAndTitle(ctx context.Context, artistID string, title string) (*model.Album, error) {
	var album model.Album

	opts := []ports.FindOption{
		ports.WithQuery(ports.NewQuery("artist_id = ? AND title = ?", artistID, title)), // Pass artistID and title as separate arguments
	}

	log.Printf("Query: artist_id = %s, title = %s", artistID, title) // Log the query parameters

	if err := r.db.FindOne(ctx, &album, opts...); err != nil {
		return nil, err
	}

	return &album, nil
}

func NewAlbumRepository(db ports.DBPort) IAlbumRepository {
	return &AlbumRepository{
		db: db,
	}
}

func (r *AlbumRepository) Create(ctx context.Context, album *model.Album) error {
	return r.db.Create(ctx, album)
}
