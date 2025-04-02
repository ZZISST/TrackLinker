package main

import (
	"log"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/adapters/db"
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
	cfg := config.LoadConfig()

	db, err := db.NewAdapter("host=localhost user=postgres password=p4ssw0rd dbname=tracks_db port=5433 sslmode=disable TimeZone=Europe/Moscow")
	if err != nil {
		log.Fatal("Database connection fail", err)
	}

	err = db.AutoMigrate(&model.Album{}, &model.TrackPreview{}, &model.Track{}, &model.MusicLink{}, &model.Artist{})
	if err != nil {
		log.Fatal("Database connection fail", err)
	}

	if err := AutoMigrateMusicService(db); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	if err := model.SeedMusicServices(db); err != nil {
		log.Fatalf("Failed to seed music services: %v", err)
	}

	s3 := s3.NewAdapter(*cfg)

	// Initialize the HTTP server
	server := httpServer.NewServer(db, s3)

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
