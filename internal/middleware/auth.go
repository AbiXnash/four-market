package middleware

import (
	"context"
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
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, r, http.StatusUnauthorized, "invalid_authorization_header", "invalid authorization header")
				return
			}

			claims, err := security.ValidateJWT(parts[1], secret)
			if err != nil {
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
