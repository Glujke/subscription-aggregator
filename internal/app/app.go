package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"subscription-aggregator/internal/config"
	"subscription-aggregator/internal/database"
	"subscription-aggregator/internal/domain"
	"subscription-aggregator/internal/handler"
	"subscription-aggregator/internal/repository/postgres"
	"subscription-aggregator/internal/service"
)

const shutdownTimeout = 10 * time.Second

// Run запускает HTTP-сервер и корректно завершает работу по сигналу.
func Run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	logger := NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	if err := database.MigrateUp(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		return fmt.Errorf("миграции: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("пул соединений: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("проверка подключения к БД: %w", err)
	}

	repo := postgres.New(pool)
	svc := service.New(repo, newNowFunc())
	router := handler.NewRouter(logger, svc)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("сервер запущен", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stopCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-stopCtx.Done():
		logger.Info("получен сигнал завершения")
	case err := <-errCh:
		return fmt.Errorf("http-сервер: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("остановка сервера: %w", err)
	}

	logger.Info("сервер остановлен")
	return nil
}

func loadConfig() (config.Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		return config.LoadFromFile(".env")
	}

	return config.Load()
}

func newNowFunc() service.NowFunc {
	loc := config.Location()
	return func() domain.MonthYear {
		now := time.Now().In(loc)
		return domain.MonthYear{Month: int(now.Month()), Year: now.Year()}
	}
}
