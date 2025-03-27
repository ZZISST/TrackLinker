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
	Environment string `env:"environment"`
	HttpPort    int    `env:"http_port"`
	GrpcPort    int    `env:"grpc_port"`
	DatabaseURI string `env:"database_uri"`
	S3Bucket    string `env:"s3_bucket"`
	S3Region    string `env:"s3_region"`
	S3Endpoint  string `env:"s3_endpoint"`
	S3AccessKey string `env:"s3_access_key"`
	S3SecretKey string `env:"s3_secret_key"`
}

func LoadConfig() *Schema {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	err := godotenv.Load(filepath.Join(currentDir, "config.yaml"))
	if err != nil {
		log.Printf("Error on load configuration file, error: %v", err)
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error on parsing configuration file, error: %v", err)
	}

	return &cfg
}

func GetConfig() *Schema {
	return &cfg
}
