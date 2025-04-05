package config

import (
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

const (
	ProductionEnv = "production"

	DatabaseTimeout = 5 * time.Second
)

var (
	cfg Schema
)

type Schema struct {
	Environment         string `env:"ENVIROMENRT"`
	SpotifyClientID     string `env:"SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret string `env:"SPOTIFY_CLIENT_SECRET"`
	ProxyURL            string `env:"PROXY_URL"`
	DatabaseURI         string `env:"DATABASE_URI"`
	S3Bucket            string `env:"S3_BUCKET"`
	S3Region            string `env:"S3_REGION"`
	S3Endpoint          string `env:"S3_ENDPOINT"`
	S3AccessKey         string `env:"S3_ACCESS_KEY"`
	S3SecretKey         string `env:"S3_SECRET_KEY"`
}

func LoadConfig(environment string) *Schema {
	_, filename, _, _ := runtime.Caller(0)

	// .env in root
	currentDir := filepath.Join(filepath.Dir(filename), "../..")
	currentDir, err := filepath.Abs(currentDir)
	if err != nil {
		log.Fatalf("Error resolving absolute path: %v", err)
	}

	if environment == "" {
		err := godotenv.Load(filepath.Join(currentDir, ".env.dev"))
		if err != nil {
			log.Printf("Error on load dev configuration file, error: %v", err)
		}
	} else {
		err := godotenv.Load(filepath.Join(currentDir, ".env.prod"))
		if err != nil {
			log.Printf("Error on load prod configuration file, error: %v", err)
		}
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error on parsing configuration file, error: %v", err)
	}

	return &cfg
}

func GetConfig() *Schema {
	return &cfg
}
