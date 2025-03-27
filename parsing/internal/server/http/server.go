package http

import (
	"log"
	// "net/http"

	"github.com/gin-gonic/gin"
	parsingHttp "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/transport/http"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/config"

	// "gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/response"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Schema
	db     ports.DBPort
	s3     ports.S3Port
}

func NewServer(db ports.DBPort, s3 ports.S3Port) *Server {
	return &Server{
		engine: gin.New(),
		cfg:    config.GetConfig(),
		db:     db,
		s3:     s3,
	}
}

func (s Server) Run() error {
	// _ = s.engine.SetTrustedProxies(nil)
	// if s.cfg.Environment == config.ProductionEnv {
	// gin.SetMode(gin.ReleaseMode)
	// }

	gin.SetMode(gin.ReleaseMode)
	log.Printf("Gin is running in %s mode", gin.Mode())

	if err := s.MapRoutes(); err != nil {
		log.Fatalf("MapRoutes Error: %v", err)
	}
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return s.engine.Run("0.0.0.0" + ":" + "8080")
}

func (s Server) GetEngine() *gin.Engine {
	return s.engine
}

func (s Server) MapRoutes() error {
	v1 := s.engine.Group("/api/v1")
	parsingHttp.Routes(v1, s.db, s.s3)
	return nil
}
