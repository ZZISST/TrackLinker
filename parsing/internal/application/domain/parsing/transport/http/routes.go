package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service/cloudservice"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service/collectionservice"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service/musicservice"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

func Routes(r *gin.RouterGroup, db ports.DBPort, s3 ports.S3Port, collection ports.CollectionPort) {
	albumRepo := repository.NewAlbumRepository(db)
	artistRepo := repository.NewArtistRepository(db)
	trackRepo := repository.NewTrackRepository(db)
	musicLinkRepo := repository.NewMusicLinkRepository(db)
	trackPreviewRepo := repository.NewTrackPreviewRepository(db)
	albumArtistRepo := repository.NewAlbumArtistRepository(db)
	genreRepo := repository.NewGenreRepository(db)

	musicRepos := repository.MusicRepos{
		TrackRepo:        trackRepo,
		AlbumRepo:        albumRepo,
		ArtistRepo:       artistRepo,
		AlbumArtistRepo:  albumArtistRepo,
		MusicLinkRepo:    musicLinkRepo,
		TrackPreviewRepo: trackPreviewRepo,
		GenreRepo:        genreRepo,
	}

	spotifyAuthSvc := musicservice.NewSpotifyAuthService()
	spotifySvc := musicservice.NewSpotifyService(spotifyAuthSvc)

	yandexSvc := musicservice.NewYandexService()
	deezerSvc := musicservice.NewDeezerService()
	appleSvc := musicservice.NewAppleService()

	services := []musicservice.IParsingMusicService{spotifySvc, yandexSvc, deezerSvc, appleSvc}

	parsingSvc := service.NewParsingService(deezerSvc, services...)

	cloudSvc := cloudservice.NewYandexCloudService(s3)

	musicSvc := service.NewMusicService(musicRepos, parsingSvc, cloudSvc)

	collectionSvc := collectionservice.NewCollectionService(musicSvc, collection)

	collectionSvc.ListenQueue()

	parsingHandler := NewParsingHandler(musicSvc)

	r.GET("/track/:id", parsingHandler.GetTrackByID)
	r.GET("/album/:id", parsingHandler.GetAlbumByID)
	r.GET("/album/:album_id/track/:track_id", parsingHandler.GetAlbumTrackByID)
	r.POST("/music/search", parsingHandler.SearchMusic)
}
