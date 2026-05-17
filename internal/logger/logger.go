package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

func Init() (slog.Handler, string) {
	var error string

	logLevel := strings.ToUpper(
		os.Getenv("LOG_LEVEL"),
	)

	var level slog.Level

	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug

	case "WARN":
		level = slog.LevelWarn

	case "ERROR":
		level = slog.LevelError

	default:
		error = "LOG_LEVEL is missing in env. Using INFO as default"
		level = slog.LevelInfo
	}

	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		AddSource:  true,
		TimeFormat: "02 Jan 2006 03:04:05 PM",
	}), error
}
