package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AbiXnash/4-market/internal/logger"
	"github.com/AbiXnash/4-market/internal/server"
)

func init() {
	handler, msg := logger.Init()
	slog.SetDefault(slog.New(handler))

	if msg != "" {
		slog.Warn(msg)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server.Start(ctx)
}
