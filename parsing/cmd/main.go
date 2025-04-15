package main

import (
	"log"
	"log/slog"
	"os"

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

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)

	if cfg.Environment == config.ProductionEnv {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)

	db, err := db.NewAdapter(cfg.DatabaseURI)
	if err != nil {
		log.Fatal("Database connection fail", err)
	}

	err = db.AutoMigrate(&model.Album{}, &model.TrackPreview{}, &model.Track{}, &model.MusicLink{}, &model.Artist{}, &model.Genre{})
	if err != nil {
		log.Fatal("Database connection fail", err)
	}

	if err := AutoMigrateMusicService(db); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	if err := SeedMusicServices(db); err != nil {
		log.Fatalf("Failed to seed music services: %v", err)
	}

	s3 := s3.NewAdapter(*cfg)

	mq, err := mq.NewAdapter(cfg.AMQPURL)
	if err != nil {
		log.Fatal("Message Queue connection fail", err)
	}

	collection := collection.NewAdapter(mq)

	// Initialize the HTTP server
	server := httpServer.NewServer(db, s3, mq, collection)

	// Run the server (this blocks the program)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
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
	return db.AutoMigrate(&model.MusicServiceDictionary{})
}

func SeedMusicServices(db *db.Adapter) error {
	services := []model.MusicServiceDictionary{
		{Service: model.Apple},
		{Service: model.Yandex},
		{Service: model.Spotify},
		{Service: model.Deezer},
		{Service: model.Unsupported},
	}

	for _, service := range services {
		if err := db.GetDB().FirstOrCreate(&service, service).Error; err != nil {
			return err
		}
	}

	return nil
}
