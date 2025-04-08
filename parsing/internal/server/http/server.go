package http

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/docs"
	_ "gitlab.com/etherealSent/tracklinker-backend/parsing/docs"
	parsingHttp "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/transport/http"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/config"
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

	docs.SwaggerInfo.Host = s.cfg.SwaggerHost

	if s.cfg.Environment == config.ProductionEnv {
		docs.SwaggerInfo.BasePath = "/"              // adjust if your API is under a sub-path
		docs.SwaggerInfo.Schemes = []string{"https"} // use https since you're using SSL
	}

	gin.SetMode(gin.ReleaseMode)
	log.Printf("Gin is running in %s mode", gin.Mode())

	if s.cfg.Environment == config.ProductionEnv {
		s.engine.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"https://tracklinker.ru", "*"},
			AllowMethods:     []string{"GET", "POST", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	if err := s.MapRoutes(); err != nil {
		log.Fatalf("MapRoutes Error: %v", err)
	}

	if s.cfg.Environment == config.ProductionEnv {
		s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("https://tracklinker.ru/swagger/doc.json")))
	} else {
		s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

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
