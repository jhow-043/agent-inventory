// Package database handles PostgreSQL connection and migration execution.
package database

import (
	"context"
	"io/fs"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

// Connect establishes a connection to PostgreSQL and verifies it with a ping.
func Connect(databaseURL string) *sqlx.DB {
	db, err := sqlx.Connect("pgx", databaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		panic(err)
	}

	slog.Info("database connected successfully")
	return db
}

// RunMigrations applies all pending database migrations from the embedded filesystem.
func RunMigrations(databaseURL string, migrationsFS fs.FS) {
	source, err := iofs.New(migrationsFS, ".")
	if err != nil {
		slog.Error("failed to create migration source", "error", err)
		panic(err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		slog.Error("failed to create migrate instance", "error", err)
		panic(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", "error", err)
		panic(err)
	}

	slog.Info("database migrations completed")
}
