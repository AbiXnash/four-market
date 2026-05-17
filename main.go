package main

import (
	"log/slog"

	"github.com/AbiXnash/4-market/internal/logger"
	"github.com/AbiXnash/4-market/internal/server"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()

	handler, msg := logger.Init()
	slog.SetDefault(slog.New(handler))

	if msg != "" {
		slog.Warn(msg)
	}
}

func main() {
	server.Start()
}
