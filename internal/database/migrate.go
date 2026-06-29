package database

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateUp применяет все доступные миграции.
func MigrateUp(databaseURL, migrationsDir string) error {
	m, err := newMigrate(databaseURL, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// MigrateDown откатывает все применённые миграции.
func MigrateDown(databaseURL, migrationsDir string) error {
	m, err := newMigrate(databaseURL, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func newMigrate(databaseURL, migrationsDir string) (*migrate.Migrate, error) {
	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("путь к миграциям: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(absDir))
	return migrate.New(sourceURL, databaseURL)
}
