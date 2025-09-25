package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	URL string `env:"DATABASE_DSN"`
}

func NewPostgresDriver(ctx context.Context, cfg *PostgresConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
