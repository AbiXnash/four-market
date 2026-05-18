package router

import (
	"net/http"
	"slices"
	"time"

	"github.com/AbiXnash/4-market/internal/config"
	"github.com/AbiXnash/4-market/internal/handler"
	"github.com/AbiXnash/4-market/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type Router struct {
	*chi.Mux
}

func New(cfg config.Config) *Router {
	r := chi.NewRouter()

	setupMiddleware(r, cfg)
	setupRoutes(r, cfg)

	return &Router{Mux: r}
}

func setupMiddleware(r *chi.Mux, cfg config.Config) {
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.CleanPath)
	r.Use(middleware.SecurityHeaders())
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(middleware.RequestBodyLimiter(cfg.MaxRequestBody))
	r.Use(chiMiddleware.AllowContentType("application/json"))
	r.Use(corsHandler(cfg.CORSAllowedOrigins))
	r.Use(middleware.CSRFProtection(cfg.CORSAllowedOrigins))
}

func setupRoutes(r *chi.Mux, cfg config.Config) {
	r.Get("/", handler.Greet)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", handler.Login)
		r.Post("/auth/register", handler.Register)
		r.Post("/auth/refresh", handler.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth([]byte(cfg.JWTSecret)))
			r.Get("/user", handler.Greet)
		})
	})
}

func corsHandler(allowedOrigins []string) func(http.Handler) http.Handler {
	hasWildcard := slices.Contains(allowedOrigins, "*")

	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID", "Accept"},
		ExposedHeaders:   []string{"X-Request-ID", "Link"},
		AllowCredentials: !hasWildcard,
		MaxAge:           300,
	})
}
