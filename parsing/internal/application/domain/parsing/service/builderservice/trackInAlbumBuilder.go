package builderservice

import "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"

type TrackInAlbumBuilder struct {
	TrackRepo        repository.ITrackRepository
	AlbumRepo        repository.IAlbumRepository
	ArtistRepo       repository.IArtistRepository
	MusicLinkRepo    repository.IMusicLinkRepository
	TrackPreviewRepo repository.ITrackPreviewRepository
	AlbumArtistRepo  repository.IAlbumArtistRepository
	GenreRepo        repository.IGenreRepository
}

func NewTrackInAlbumBuilder() *TrackInAlbumBuilder {
	return &TrackInAlbumBuilder{}
}
