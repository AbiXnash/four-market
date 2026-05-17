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
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		next.ServeHTTP(rw, r)

		rec := slog.NewRecord(
			time.Now(),
			slog.LevelInfo,
			fmt.Sprintf("%s | %s | [%d %s] | %s", r.Method, r.URL.Path, rw.status, http.StatusText(rw.status), time.Since(start).Round(time.Microsecond)),
			0,
		)
		_ = slog.Default().Handler().Handle(r.Context(), rec)
	})
}
