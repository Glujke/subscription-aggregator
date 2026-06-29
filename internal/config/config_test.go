package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"subscription-aggregator/internal/config"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MIGRATIONS_PATH", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr: получили %q, ожидали %q", cfg.HTTPAddr, ":8080")
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel: получили %q, ожидали %q", cfg.LogLevel, "info")
	}
	if cfg.MigrationsPath != "migrations" {
		t.Fatalf("MigrationsPath: получили %q, ожидали %q", cfg.MigrationsPath, "migrations")
	}
}

func TestLoad_requiredDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("ожидалась ошибка при пустом DATABASE_URL")
	}
}

func TestLoad_customValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom:5432/db?sslmode=disable")
	t.Setenv("HTTP_ADDR", ":9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("MIGRATIONS_PATH", "deploy/migrations")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if cfg.DatabaseURL != "postgres://custom:5432/db?sslmode=disable" {
		t.Fatalf("DatabaseURL: получили %q", cfg.DatabaseURL)
	}
	if cfg.HTTPAddr != ":9000" {
		t.Fatalf("HTTPAddr: получили %q", cfg.HTTPAddr)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel: получили %q", cfg.LogLevel)
	}
	if cfg.MigrationsPath != "deploy/migrations" {
		t.Fatalf("MigrationsPath: получили %q", cfg.MigrationsPath)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `HTTP_ADDR=:3000
DATABASE_URL=postgres://file:5432/db?sslmode=disable
LOG_LEVEL=warn
MIGRATIONS_PATH=custom/migrations
`
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("не удалось записать .env: %v", err)
	}

	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MIGRATIONS_PATH", "")

	cfg, err := config.LoadFromFile(envPath)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if cfg.HTTPAddr != ":3000" {
		t.Fatalf("HTTPAddr: получили %q", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://file:5432/db?sslmode=disable" {
		t.Fatalf("DatabaseURL: получили %q", cfg.DatabaseURL)
	}
	if cfg.LogLevel != "warn" {
		t.Fatalf("LogLevel: получили %q", cfg.LogLevel)
	}
	if cfg.MigrationsPath != "custom/migrations" {
		t.Fatalf("MigrationsPath: получили %q", cfg.MigrationsPath)
	}
}

func TestLocation_moscow(t *testing.T) {
	t.Parallel()

	loc := config.Location()
	if loc.String() != "Europe/Moscow" {
		t.Fatalf("получили %q, ожидали Europe/Moscow", loc.String())
	}
}
