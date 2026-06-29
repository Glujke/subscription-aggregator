package postgres_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"subscription-aggregator/internal/database"
	repopostgres "subscription-aggregator/internal/repository/postgres"
)

func setupRepository(t *testing.T) (*repopostgres.Repository, uuid.UUID) {
	t.Helper()

	ctx := context.Background()

	pg, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("subscriptions_repo_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("не удалось поднять контейнер: %v", err)
	}
	t.Cleanup(func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Fatalf("не удалось остановить контейнер: %v", err)
		}
	})

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("не удалось получить строку подключения: %v", err)
	}

	if err := waitForPostgres(ctx, connStr); err != nil {
		t.Fatalf("postgres не готов: %v", err)
	}

	if err := database.MigrateUp(connStr, migrationsPath(t)); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("не удалось создать пул: %v", err)
	}
	t.Cleanup(pool.Close)

	return repopostgres.New(pool), uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
}

func migrationsPath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("не удалось определить путь к тесту")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")
}

func waitForPostgres(ctx context.Context, connStr string) error {
	var lastErr error

	for range 30 {
		pool, err := pgxpool.New(ctx, connStr)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		lastErr = pool.Ping(ctx)
		pool.Close()
		if lastErr == nil {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return lastErr
}
