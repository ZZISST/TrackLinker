package main

import (
	"log"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/adapters/db"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/adapters/mq"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/adapters/parsing"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model/modelMusic"
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

	err = db.AutoMigrate(&modelMusic.Album{}, &modelMusic.TrackPreview{}, &modelMusic.Track{}, &modelMusic.MusicLink{}, &modelMusic.Artist{}, model.Collection{}, model.AddedContent{}, model.AddedContentType{}, model.ContentTag{}, model.Listening{}, model.ReferralClick{}, model.ReferralLink{})
	if err != nil {
		log.Fatal("Database connection fail", err)
	}

	if err := AutoMigrateMusicService(db); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	if err := SeedMusicServices(db); err != nil {
		log.Fatalf("Failed to seed music services: %v", err)
	}

	mq, err := mq.NewAdapter(cfg.AMQPURL)
	if err != nil {
		log.Fatal("Message Queue connection fail", err)
	}

	parsing := parsing.NewAdapter(mq)

	// Initialize the HTTP server
	server := httpServer.NewServer(db, parsing)

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
	return db.AutoMigrate(&modelMusic.MusicServiceDictionary{})
}

func SeedMusicServices(db *db.Adapter) error {
	services := []modelMusic.MusicServiceDictionary{
		{Service: modelMusic.Apple},
		{Service: modelMusic.Yandex},
		{Service: modelMusic.Spotify},
		{Service: modelMusic.Deezer},
		{Service: modelMusic.Unsupported},
	}

	for _, service := range services {
		if err := db.GetDB().FirstOrCreate(&service, service).Error; err != nil {
			return err
		}
	}

	return nil
}
