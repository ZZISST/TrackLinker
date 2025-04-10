package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/repository"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/service"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

func Routes(r *gin.RouterGroup, db ports.DBPort, parsing ports.ParsingPort) {

	addedContentRepo := repository.NewAddedContentRepository(db)
	artistRepo := repository.NewArtistRepository(db)
	collectionRepo := repository.NewCollectionRepository(db)
	listeningRepo := repository.NewListeningRepository(db)
	trackRepo := repository.NewTrackRepository(db)

	parsingSvc := service.NewParsingService(parsing)
	listeningSvc := service.NewListeningService(listeningRepo, addedContentRepo, collectionRepo, trackRepo, artistRepo, parsingSvc)

	collectionHandler := NewCollectionHandler(listeningSvc)

	actionRoute := r.Group("/action")
	{
		actionRoute.POST("/listen", collectionHandler.MakeListening)
	}
}
