package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/repository"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

func Routes(r *gin.RouterGroup, db ports.DBPort, s3 ports.S3Port) {
	albumRepo := repository.NewAlbumRepository(db)
	artistRepo := repository.NewArtistRepository(db)
	trackRepo := repository.NewTrackRepository(db)
	musicLinkRepo := repository.NewMusicLinkRepository(db)
	trackPreviewRepo := repository.NewTrackPreviewRepository(db)

	spotifyAuthSvc := service.NewSpotifyAuthService()
	spotifySvc := service.NewSpotifyService(spotifyAuthSvc)

	yandexSvc := service.NewYandexService()
	deezerSvc := service.NewDeezerService()
	appleSvc := service.NewAppleService()

	services := []service.IParsingMusicService{spotifySvc, yandexSvc, deezerSvc, appleSvc}

	parsingSvc := service.NewParsingService(deezerSvc, services...)

	cloudSvc := service.NewYandexCloudService(s3)

	musicSvc := service.NewMusicService(trackRepo, artistRepo, albumRepo, musicLinkRepo, trackPreviewRepo, parsingSvc, cloudSvc)

	parsingHandler := NewParsingHandler(musicSvc)

	musicRoute := r.Group("/music")
	{
		musicRoute.GET("/:id", parsingHandler.GetMusicByID)
		musicRoute.POST("/track", parsingHandler.FindMusic)
	}
}
