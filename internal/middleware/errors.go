package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/AbiXnash/4-market/internal/response"
)

func RecoverJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				slog.Error("panic recovered",
					"request_id", r.Header.Get("X-Request-ID"),
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rec,
					"stack", string(debug.Stack()),
				)

				if r.Header.Get("Connection") == "Upgrade" {
					return
				}

				response.InternalError(w, r)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func AllowJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		contentType := strings.TrimSpace(strings.ToLower(r.Header.Get("Content-Type")))
		for i := 0; i < len(contentType); i++ {
			if contentType[i] == ';' {
				contentType = strings.TrimSpace(contentType[:i])
				break
			}
		}

		if contentType == "application/json" {
			next.ServeHTTP(w, r)
			return
		}

		response.UnsupportedMediaType(w, r)
	})
}
