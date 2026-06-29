package database_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"subscription-aggregator/internal/database"
)

func TestMigrations_upAndDown(t *testing.T) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("subscriptions_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
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

	migrationsDir := migrationsPath(t)

	if err := database.MigrateUp(connStr, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("не удалось создать пул: %v", err)
	}
	t.Cleanup(pool.Close)

	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'subscriptions'
		)`).Scan(&exists)
	if err != nil {
		t.Fatalf("проверка таблицы: %v", err)
	}
	if !exists {
		t.Fatal("таблица subscriptions не найдена после migrate up")
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO subscriptions (service_name, price, user_id, start_date)
		VALUES ('Yandex Plus', 400, '60601fee-2bf1-4721-ae6f-7636e79a0cba', '2025-07-01')
	`)
	if err != nil {
		t.Fatalf("smoke insert: %v", err)
	}

	if err := database.MigrateDown(connStr, migrationsDir); err != nil {
		t.Fatalf("migrate down: %v", err)
	}

	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'subscriptions'
		)`).Scan(&exists)
	if err != nil {
		t.Fatalf("проверка таблицы после down: %v", err)
	}
	if exists {
		t.Fatal("таблица subscriptions осталась после migrate down")
	}
}

func migrationsPath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("не удалось определить путь к тесту")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
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
