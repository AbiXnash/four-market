package router

import (
	"net/http"
	"slices"
	"time"

	"github.com/AbiXnash/4-market/internal/config"
	"github.com/AbiXnash/4-market/internal/handler"
	"github.com/AbiXnash/4-market/internal/middleware"
	"github.com/AbiXnash/4-market/internal/response"
	"github.com/AbiXnash/4-market/web"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type Router struct {
	*chi.Mux
}

func New(cfg config.Config, authH *handler.AuthHandler, userH *handler.UserHandler) *Router {
	r := chi.NewRouter()

	setupMiddleware(r, cfg)
	setupRoutes(r, cfg, authH, userH)

	return &Router{Mux: r}
}

func setupMiddleware(r *chi.Mux, cfg config.Config) {
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.RecoverJSON)
	r.Use(chiMiddleware.CleanPath)
	r.Use(middleware.SecurityHeaders())
	r.Use(httprate.Limit(
		100,
		1*time.Minute,
		httprate.WithKeyByIP(),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			response.TooManyRequests(w, r)
		}),
		httprate.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			response.Error(w, r, http.StatusInternalServerError, "rate_limiter_error", "failed to evaluate rate limit")
		}),
	))
	r.Use(middleware.RequestBodyLimiter(cfg.MaxRequestBody))
	r.Use(corsHandler(cfg.CORSAllowedOrigins))
	r.Use(middleware.CSRFProtection(cfg.CORSAllowedOrigins))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, r)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.MethodNotAllowed(w, r)
	})
}

func setupRoutes(r *chi.Mux, cfg config.Config, authH *handler.AuthHandler, userH *handler.UserHandler) {
	r.Get("/", templ.Handler(web.LoginForm()).ServeHTTP)
	r.Post("/login", authH.FormLogin)
	r.Get("/register", authH.FormRegister)
	r.Post("/register", authH.FormRegister)
	r.Post("/logout", authH.FormLogout)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/refresh", authH.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth([]byte(cfg.JWTSecret)))
			r.Get("/user", userH.GetCurrentUser)
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
