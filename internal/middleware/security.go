package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/AbiXnash/4-market/internal/response"
	"github.com/unrolled/secure"
)

type ctxKey string

const RequestIDKey ctxKey = "request_id"

func SecurityHeaders() func(http.Handler) http.Handler {
	sec := secure.New(secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		SSLRedirect:           false,
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; object-src 'none'; img-src 'self' data:;",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=(), payment=(), usb=()",
	})
	return sec.Handler
}

func RequestBodyLimiter(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				slog.Error("failed to generate request ID", "error", err)
				next.ServeHTTP(w, r)
				return
			}
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		r.Header.Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CSRFProtection(allowedOrigins []string) func(http.Handler) http.Handler {
	parsed := parseOrigins(allowedOrigins)
	allowAll := len(parsed) == 0 || (len(parsed) == 1 && parsed[0].Host == "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowAll {
				switch r.Method {
				case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
					origin := r.Header.Get("Origin")
					referer := r.Header.Get("Referer")

					originStr := origin
					if originStr == "" {
						originStr = referer
					}

					if originStr != "" && !isSameOrigin(originStr, r) && !originAllowed(originStr, parsed) && !refererAllowed(originStr, parsed) {
						response.Forbidden(w, r)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isSameOrigin(rawurl string, r *http.Request) bool {
	u, err := url.Parse(rawurl)
	if err != nil || u.Host == "" {
		return false
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return u.Host == r.Host && u.Scheme == scheme
}

func parseOrigins(origins []string) []*url.URL {
	parsed := make([]*url.URL, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "*" || o == "" {
			parsed = append(parsed, &url.URL{Scheme: "*", Host: "*"})
			continue
		}
		u, err := url.Parse(o)
		if err == nil && u.Host != "" {
			parsed = append(parsed, u)
		}
	}
	return parsed
}

func originAllowed(origin string, allowed []*url.URL) bool {
	o, err := url.Parse(origin)
	if err != nil || o.Host == "" {
		return false
	}
	for _, a := range allowed {
		if a.Host == o.Host && (a.Scheme == o.Scheme || a.Scheme == "*") {
			return true
		}
	}
	return false
}

func refererAllowed(referer string, allowed []*url.URL) bool {
	r, err := url.Parse(referer)
	if err != nil || r.Host == "" {
		return false
	}
	for _, a := range allowed {
		if a.Host == r.Host && (a.Scheme == r.Scheme || a.Scheme == "*") {
			return true
		}
	}
	return false
}
