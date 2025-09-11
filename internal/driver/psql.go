package driver

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/config/db"
	"github.com/jackc/pgx/v5"
)

func NewPostgresDriver(ctx context.Context, cfg *db.PostgresConfig) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, cfg.URL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
