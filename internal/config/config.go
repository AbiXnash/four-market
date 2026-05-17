package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

func Load() Config {

	root, err := os.Getwd()

	if err != nil {
		slog.Error("failed to get working directory")
		os.Exit(1)
	}

	envPath := filepath.Join(root, ".env")

	err = godotenv.Load(envPath)

	if err != nil {
		slog.Error(".env file not loaded")
		os.Exit(1)
	}

	port := os.Getenv("PORT")

	if port == "" {
		slog.Error("PORT is missing")
		os.Exit(1)
	}

	return Config{
		Port: port,
	}
}
