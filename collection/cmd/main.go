package main

import (
	"log"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/adapters/db"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model"
	httpServer "gitlab.com/etherealSent/tracklinker-backend/collection/internal/server/http"
	"gitlab.com/etherealSent/tracklinker-backend/collection/pkg/config"
)

//	@title			TrackLinker Swagger API
//	@version		2.0
//	@description	Swagger API for GoShop.
//	@termsOfService	http://swagger.io/terms/

// @BasePath	/api/v1
func main() {
	cfg := config.LoadConfig("")

	db, err := db.NewAdapter(cfg.DatabaseURI)
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

	// Initialize the HTTP server
	server := httpServer.NewServer(db)

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
