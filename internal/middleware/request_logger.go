package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Round(time.Microsecond)
		reqID, _ := r.Context().Value(RequestIDKey).(string)

		slog.LogAttrs(r.Context(), slog.LevelInfo,
			fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.status),
			slog.Duration("duration", duration),
			slog.String("remote_ip", r.RemoteAddr),
			slog.String("request_id", reqID),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}
