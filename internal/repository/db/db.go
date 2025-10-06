package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseRepository struct {
	db *pgxpool.Pool
}

type DatabaseRepository interface {
	StartTransaction(context.Context) (pgx.Tx, error)
	IntrospectTransaction(ctx context.Context, tx pgx.Tx, err error)
	Replace(ctx context.Context, tx pgx.Tx, name string, value float64) error
	Add(ctx context.Context, tx pgx.Tx, name string, value int64) error
	ReadCounter(string) (int64, error)
	ReadGauge(string) (float64, error)
	ReadAll() map[string]string
	Ping() bool
}

func NewDatabaseRepository(pool *pgxpool.Pool) DatabaseRepository {
	return &databaseRepository{
		db: pool,
	}
}
