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
	Port               string
	CORSAllowedOrigins []string
	MaxRequestBody     int64
	JWTSecret          string
	JWTAccessTTL       int // minutes
	JWTRefreshTTL      int // minutes
	TLSEnabled         bool
	TLSCertFile        string
	TLSKeyFile         string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
}

func Load() Config {
	root, err := os.Getwd()
	if err != nil {
		slog.Error("failed to get working directory")
		os.Exit(1)
	}

	// .env is local-only; production uses real env vars
	envPath := filepath.Join(root, ".env")
	if err := godotenv.Load(envPath); err != nil {
		slog.Debug("no .env file found, using system environment")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "2323"
	}

	corsOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var corsOrigins []string
	if corsOriginsStr != "" {
		corsOrigins = strings.Split(corsOriginsStr, ",")
	} else if os.Getenv("PORT") == "" {
		// no .env and no system env → dev defaults
		corsOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	} else {
		// production without explicit setting — restrict
		corsOrigins = []string{}
	}

	maxBodyStr := os.Getenv("MAX_REQUEST_BODY")
	maxBody := int64(1 << 20)
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
		jwtRefreshTTL = 10080
	}

	tlsEnabled, _ := strconv.ParseBool(os.Getenv("TLS_ENABLED"))

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	return Config{
		Port:               port,
		CORSAllowedOrigins: corsOrigins,
		MaxRequestBody:     maxBody,
		JWTSecret:          jwtSecret,
		JWTAccessTTL:       jwtAccessTTL,
		JWTRefreshTTL:      jwtRefreshTTL,
		TLSEnabled:         tlsEnabled,
		TLSCertFile:        os.Getenv("TLS_CERT_FILE"),
		TLSKeyFile:         os.Getenv("TLS_KEY_FILE"),
		RedisAddr:          redisAddr,
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisDB:            redisDB,
	}
}
