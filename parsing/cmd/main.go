package main

import (
	"github.com/gookit/slog"
	// "github.com/gookit/slog/handler"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/adapters/collection"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/adapters/db"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/adapters/mq"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/adapters/s3"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
	httpServer "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/server/http"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/config"
)

//	@title			TrackLinker Swagger API
//	@version		2.0
//	@description	Swagger API for GoShop.
//	@termsOfService	http://swagger.io/terms/

// @BasePath	/api/v1
func main() {
	cfg := config.LoadConfig("")

	// opts := &slog.HandlerOptions{
	// 	Level: slog.LevelDebug,
	// }

	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})

	if cfg.Environment == config.ProductionEnv {
		slog.SetFormatter(slog.NewJSONFormatter())
	}

	db, err := db.NewAdapter(cfg.DatabaseURI)
	if err != nil {
		slog.Fatal("Database connection fail", err)
	}

	err = db.AutoMigrate(&model.Album{}, &model.TrackPreview{}, &model.Track{}, &model.MusicLink{}, &model.Artist{}, &model.Genre{})
	if err != nil {
		slog.Fatal("Database connection fail", err)
	}

	if err := AutoMigrateMusicService(db); err != nil {
		slog.Fatalf("Failed to migrate: %v", err)
	}

	s3 := s3.NewAdapter(*cfg)

	mq, err := mq.NewAdapter(cfg.AMQPURL)
	if err != nil {
		slog.Fatal("Message Queue connection fail", err)
	}

	collection := collection.NewAdapter(mq)

	server := httpServer.NewServer(db, s3, mq, collection)

	if err := server.Run(); err != nil {
		slog.Fatalf("Failed to run server: %v", err)
	}

}

func AutoMigrateMusicService(db *db.Adapter) error {
	// Create the ENUM type if it doesn't exist
	if err := db.GetDB().Exec(`
        DO $$
        BEGIN
            IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'music_service') THEN
                CREATE TYPE music_service AS ENUM ('Apple', 'Yandex', 'Spotify', 'Deezer', 'Unsupported');
            END IF;
        END$$;
    `).Error; err != nil {
		return err
	}

	// AutoMigrate the table
	if err := db.AutoMigrate(&model.MusicServiceDictionary{}); err != nil {
		return err
	}

	// Seed initial data
	services := []model.MusicServiceDictionary{
		{Service: model.Apple},
		{Service: model.Yandex},
		{Service: model.Spotify},
		{Service: model.Deezer},
	}

	for _, service := range services {
		if err := db.GetDB().FirstOrCreate(&service, service).Error; err != nil {
			return err
		}
	}

	return nil
}
