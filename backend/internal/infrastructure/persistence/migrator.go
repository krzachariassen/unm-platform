package persistence

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	pgxstdlib "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations applies all pending up migrations using golang-migrate
// with the SQL files embedded at compile time.
func RunMigrations(databaseURL string) error {
	cc, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse database URL: %w", err)
	}

	db := pgxstdlib.OpenDB(*cc)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}

	src, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}
