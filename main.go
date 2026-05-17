package main

import (
	"log/slog"

	"github.com/AbiXnash/4-market/internal/logger"
	"github.com/AbiXnash/4-market/internal/server"
)

func init() {
	handler := logger.Init()
	slog.SetDefault(slog.New(handler))
}

func main() {
	server.Start()
}
