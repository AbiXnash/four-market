package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/AbiXnash/4-market/internal/config"
	"github.com/AbiXnash/4-market/internal/handler"
	"github.com/AbiXnash/4-market/internal/middleware"
)

func Start() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Greet)

	slog.Info("Server started", "port", cfg.Port)

	http.ListenAndServe(
		fmt.Sprintf(":%s", cfg.Port),
		middleware.LoggerMiddleware(mux),
	)
}
