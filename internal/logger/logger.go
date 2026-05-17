package logger

import (
	"fmt"
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
	case "":
		error = "LOG_LEVEL is missing in env. Using INFO as default"
		level = slog.LevelInfo

	case "DEBUG":
		level = slog.LevelDebug

	case "WARN":
		level = slog.LevelWarn

	case "ERROR":
		level = slog.LevelError

	default:
		error = fmt.Sprintf("LOG_LEVEL invalid. [%v]", logLevel)
		level = slog.LevelInfo
	}

	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		AddSource:  true,
		TimeFormat: "02 Jan 2006 03:04:05 PM",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey && len(groups) == 0 {
				switch l := a.Value.Any().(slog.Level); l {
				case slog.LevelDebug:
					return slog.String(slog.LevelKey, "DEBUG")
				case slog.LevelInfo:
					return tint.Attr(10, slog.String(slog.LevelKey, "INFO"))
				case slog.LevelWarn:
					return tint.Attr(11, slog.String(slog.LevelKey, "WARN"))
				case slog.LevelError:
					return tint.Attr(9, slog.String(slog.LevelKey, "ERROR"))
				default:
					return a
				}
			}
			return a
		},
	}), error
}
