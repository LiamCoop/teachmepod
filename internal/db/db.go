// Package db opens the Postgres pool and applies schema migrations.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"teachmepod/migrations"
)

// URL returns the connection string from DATABASE_URL.
func URL() (string, error) {
	u := os.Getenv("DATABASE_URL")
	if u == "" {
		return "", errors.New("DATABASE_URL is not set")
	}
	return u, nil
}

// Open connects to Postgres and verifies the connection.
func Open(ctx context.Context) (*pgxpool.Pool, error) {
	u, err := URL()
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

// Migrate applies all pending migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()
	return migrate(ctx, sqlDB)
}

func migrate(ctx context.Context, sqlDB *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	results, err := provider.Up(ctx)
	for _, r := range results {
		fmt.Printf("migrated %s (%s)\n", r.Source.Path, r.Duration)
	}
	return err
}
