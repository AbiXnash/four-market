package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/AbiXnash/4-market/internal/config"
	"github.com/AbiXnash/4-market/internal/handler"
	redisStore "github.com/AbiXnash/4-market/internal/redis"
	"github.com/AbiXnash/4-market/internal/repository"
	"github.com/AbiXnash/4-market/internal/router"
	"github.com/AbiXnash/4-market/internal/service"
)

func Start(ctx context.Context) {
	cfg := config.Load()

	rstore, err := redisStore.Connect(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Warn("redis not available, continuing without it", "error", err)
		rstore = &redisStore.Store{}
	} else {
		slog.Info("redis connected", "addr", cfg.RedisAddr)
	}
	defer rstore.Close()

	userRepo := repository.NewUserRepo()
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authH := handler.NewAuthHandler(authSvc, rstore, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL, cfg.TLSEnabled)
	userH := handler.NewUserHandler(userRepo)

	r := router.New(cfg, authH, userH)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}

	go func() {
		slog.Info("Server started", "port", cfg.Port, "tls", cfg.TLSEnabled)

		var err error
		if cfg.TLSEnabled {
			err = srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	<-ctx.Done()

	slog.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server stopped")
}
