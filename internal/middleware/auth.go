package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AbiXnash/4-market/internal/response"
	"github.com/AbiXnash/4-market/internal/security"
)

type ctxKeyUser string

const UserKey ctxKeyUser = "user"

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					tokenString = parts[1]
				}
			}

			if tokenString == "" {
				if cookie, err := r.Cookie("access_token"); err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				slog.Warn("auth: no token provided", "path", r.URL.Path)
				response.Unauthorized(w, r)
				return
			}

			claims, err := security.ValidateJWT(tokenString, secret)
			if err != nil {
				slog.Warn("auth: invalid token", "path", r.URL.Path, "error", err)
				response.Unauthorized(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserKey).(*security.Claims)
			if !ok || claims.Role != role {
				response.Forbidden(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
