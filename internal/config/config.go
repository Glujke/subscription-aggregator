package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrEmptyDatabaseURL = errors.New("DATABASE_URL обязателен")
	ErrInvalidLogLevel  = errors.New("недопустимый LOG_LEVEL")
)

// Config — параметры запуска сервиса.
type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	LogLevel       string
	MigrationsPath string
}

// Load читает конфигурацию из переменных окружения процесса.
func Load() (Config, error) {
	return loadFromEnv()
}

// LoadFromFile загружает .env-файл и читает конфигурацию из окружения.
func LoadFromFile(path string) (Config, error) {
	if err := godotenv.Overload(path); err != nil {
		return Config{}, fmt.Errorf("загрузка .env: %w", err)
	}
	return loadFromEnv()
}

// Location возвращает таймзону для расчёта текущего месяца.
func Location() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.UTC
	}
	return loc
}

func loadFromEnv() (Config, error) {
	cfg := Config{
		HTTPAddr:       envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL:    strings.TrimSpace(os.Getenv("DATABASE_URL")),
		LogLevel:       envOrDefault("LOG_LEVEL", "info"),
		MigrationsPath: envOrDefault("MIGRATIONS_PATH", "migrations"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, ErrEmptyDatabaseURL
	}

	if !isValidLogLevel(cfg.LogLevel) {
		return Config{}, ErrInvalidLogLevel
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func isValidLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
