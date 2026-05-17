package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/AbiXnash/4-market/internal/app"
	"github.com/AbiXnash/4-market/internal/config"
)

func Start() {
	cfg := config.Load()

	mux := app.NewRouter()

	slog.Info("Server started", "port", cfg.Port)

	http.ListenAndServe(
		fmt.Sprintf(":%s", cfg.Port),
		mux,
	)
}
