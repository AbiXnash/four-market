package app

import (
	"github.com/AbiXnash/4-market/internal/handler"
	"github.com/AbiXnash/4-market/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware)

	r.Get("/", handler.Greet)

	return r
}
