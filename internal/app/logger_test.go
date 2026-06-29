package app_test

import (
	"context"
	"log/slog"
	"testing"

	"subscription-aggregator/internal/app"
)

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{name: "debug", input: "debug", want: slog.LevelDebug},
		{name: "info", input: "info", want: slog.LevelInfo},
		{name: "warn", input: "warn", want: slog.LevelWarn},
		{name: "error", input: "error", want: slog.LevelError},
		{name: "неизвестный", input: "verbose", want: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := app.ParseLogLevel(tt.input); got != tt.want {
				t.Fatalf("получили %v, ожидали %v", got, tt.want)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	logger := app.NewLogger("error")
	if !logger.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("уровень error должен быть включён")
	}
	if logger.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("уровень info должен быть выключен")
	}
}
