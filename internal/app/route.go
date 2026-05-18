package app

import (
	"time"

	"github.com/AbiXnash/4-market/internal/config"
	"github.com/AbiXnash/4-market/internal/handler"
	"github.com/AbiXnash/4-market/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func NewRouter(cfg config.Config) *chi.Mux {
	r := chi.NewRouter()

	// --- core infrastructure ---
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.LoggerMiddleware)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.CleanPath)

	// --- security headers ---
	r.Use(middleware.NewSecurityHandler())

	// --- rate limiting (global IP-based) ---
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// --- body size limit ---
	r.Use(middleware.RequestBodyLimiter(cfg.MaxRequestBody))

	// --- content negotiation ---
	r.Use(chiMiddleware.AllowContentType("application/json"))

	// --- CORS ---
	hasWildcard := false
	for _, o := range cfg.CORSAllowedOrigins {
		if o == "*" {
			hasWildcard = true
			break
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID", "Accept"},
		ExposedHeaders:   []string{"X-Request-ID", "Link"},
		AllowCredentials: !hasWildcard,
		MaxAge:           300,
	}))

	// --- CSRF protection (defense-in-depth) ---
	r.Use(middleware.CSRFProtection(cfg.CORSAllowedOrigins))

	// --- API v1 routes ---
	r.Route("/api/v1", func(r chi.Router) {

		// public endpoints
		r.Post("/auth/login", handler.Login)
		r.Post("/auth/register", handler.Register)
		r.Post("/auth/refresh", handler.Refresh)

		// authenticated endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware([]byte(cfg.JWTSecret)))
			r.Get("/user", handler.Greet)
		})
	})

	// legacy root
	r.Get("/", handler.Greet)

	return r
}
