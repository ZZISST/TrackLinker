package repository

import (
	"context"
	"net/url"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type IAlbumRepository interface {
	Create(ctx context.Context, album *model.Album) error
	GetAlbumByArtistsIDAndTitle(ctx context.Context, artistIDs []string, title string) (*model.Album, error)
	GetAlbumByMusicLink(ctx context.Context, musicLink url.URL) (*model.Album, error)
	GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error)
}

type AlbumRepository struct {
	db ports.DBPort
}

func (r *AlbumRepository) GetAlbumByID(ctx context.Context, albumID string) (*model.Album, error) {
	var album model.Album

	if err := r.db.FindOne(ctx, &album, ports.WithQuery(ports.NewQuery("id = ?", albumID))); err != nil {
		return nil, err
	}

	return &album, nil
}

func (r *AlbumRepository) GetAlbumByMusicLink(ctx context.Context, musicLink url.URL) (*model.Album, error) {
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

func (r *AlbumRepository) GetAlbumByArtistsIDAndTitle(ctx context.Context, artistIDs []string, title string) (*model.Album, error) {
	var album model.Album

	opts := []ports.FindOption{
		ports.WithJoin("JOIN album_artists ON album_artists.album_id = albums.id"),
		ports.WithQuery(ports.NewQuery("albums.title = ? AND album_artists.artist_id IN ?", title, artistIDs)),
		ports.WithGroup("albums.id"),
		ports.WithHaving(ports.NewQuery("COUNT(DISTINCT album_artists.artist_id) = ?", len(artistIDs))),
	}

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
