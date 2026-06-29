package app

import (
	"log/slog"
	"os"
	"strings"
)

// ParseLogLevel преобразует строковый уровень логирования в slog.Level.
func ParseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewLogger создаёт логгер приложения.
func NewLogger(level string) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: ParseLogLevel(level),
	}))
}
