package repository

type MusicRepos struct {
	TrackRepo        ITrackRepository
	AlbumRepo        IAlbumRepository
	ArtistRepo       IArtistRepository
	MusicLinkRepo    IMusicLinkRepository
	TrackPreviewRepo ITrackPreviewRepository
	AlbumArtistRepo  IAlbumArtistRepository
	GenreRepo        IGenreRepository
}
