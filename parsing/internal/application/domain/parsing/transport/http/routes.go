package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
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

	spotifyAuthSvc := service.NewSpotifyAuthService()
	spotifySvc := service.NewSpotifyService(spotifyAuthSvc)

	yandexSvc := service.NewYandexService()
	deezerSvc := service.NewDeezerService()
	appleSvc := service.NewAppleService()

	services := []service.IParsingMusicService{spotifySvc, yandexSvc, deezerSvc, appleSvc}

	parsingSvc := service.NewParsingService(deezerSvc, services...)

	cloudSvc := service.NewYandexCloudService(s3)

	musicSvc := service.NewMusicService(trackRepo, artistRepo, albumRepo, musicLinkRepo, trackPreviewRepo, albumArtistRepo, genreRepo, parsingSvc, cloudSvc)

	collectionSvc := service.NewCollectionService(musicSvc, collection)

	collectionSvc.ListenQueue()

	parsingHandler := NewParsingHandler(musicSvc)

	r.GET("/track/:id", parsingHandler.GetTrackByID)
	r.GET("/album/:id", parsingHandler.GetAlbumByID)
	r.GET("/album/:album_id/track/:track_id", parsingHandler.GetAlbumTrackByID)
	r.POST("/music/search", parsingHandler.SearchMusic)
}
