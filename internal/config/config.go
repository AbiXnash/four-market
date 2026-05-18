package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	CORSAllowedOrigins []string
	MaxRequestBody    int64
	JWTSecret         string
	JWTAccessTTL      int // minutes
	JWTRefreshTTL     int // minutes
	TLSEnabled        bool
	TLSCertFile       string
	TLSKeyFile        string
}

func Load() Config {
	root, err := os.Getwd()

	if err != nil {
		slog.Error("failed to get working directory")
		os.Exit(1)
	}

	envPath := filepath.Join(root, ".env")

	slog.Debug("Looking for env in ", "path", envPath)

	err = godotenv.Load(envPath)

	if err != nil {
		slog.Error(".env file not loaded")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		slog.Error("PORT is missing")
		os.Exit(1)
	}

	corsOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var corsOrigins []string
	if corsOriginsStr != "" {
		corsOrigins = strings.Split(corsOriginsStr, ",")
	} else {
		corsOrigins = []string{"*"}
	}

	maxBodyStr := os.Getenv("MAX_REQUEST_BODY")
	maxBody := int64(1 << 20) // 1MB
	if maxBodyStr != "" {
		if v, err := strconv.ParseInt(maxBodyStr, 10, 64); err == nil && v > 0 {
			maxBody = v
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Warn("JWT_SECRET not set, using insecure default")
		jwtSecret = "change-me-in-production"
	}

	jwtAccessTTL, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TTL"))
	if jwtAccessTTL <= 0 {
		jwtAccessTTL = 15
	}

	jwtRefreshTTL, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TTL"))
	if jwtRefreshTTL <= 0 {
		jwtRefreshTTL = 10080 // 7 days
	}

	tlsEnabled, _ := strconv.ParseBool(os.Getenv("TLS_ENABLED"))

	return Config{
		Port:              port,
		CORSAllowedOrigins: corsOrigins,
		MaxRequestBody:    maxBody,
		JWTSecret:         jwtSecret,
		JWTAccessTTL:      jwtAccessTTL,
		JWTRefreshTTL:     jwtRefreshTTL,
		TLSEnabled:        tlsEnabled,
		TLSCertFile:       os.Getenv("TLS_CERT_FILE"),
		TLSKeyFile:        os.Getenv("TLS_KEY_FILE"),
	}
}
