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
	Environment string `env:"ENVIROMENRT"`
	DatabaseURI string `env:"DATABASE_URI"`
	SwaggerHost string `env:"SWAGGER_HOST"`
	AMQPURL     string `env:"AMQP_URL"`
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
