package app

import (
	"time"

	"github.com/AbiXnash/4-market/internal/handler"
	"github.com/AbiXnash/4-market/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Recoverer)
	/// r.Use(chiMiddleware.Logger)
	r.Use(middleware.LoggerMiddleware)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.CleanPath)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Use(chiMiddleware.AllowContentType("application/json"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Get("/", handler.Greet)

	return r
}
