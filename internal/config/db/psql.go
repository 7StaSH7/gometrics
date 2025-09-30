package db

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresConfig struct {
	URL string `env:"DATABASE_DSN"`
}

type queryTracer struct {
	log *zap.SugaredLogger
}

func (tracer *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	tracer.log.Infow("Executing command", "sql", data.SQL, "args", data.Args)
	return ctx
}

func (tracer *queryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

func NewPostgresDriver(ctx context.Context, cfg *PostgresConfig) (*pgxpool.Pool, error) {
	if err := autoMigrate(cfg.URL); err != nil {
		return nil, err
	}

	conf, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, err
	}

	conf.ConnConfig.Tracer = &queryTracer{
		log: logger.Log.Sugar(),
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func autoMigrate(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}

		return err
	}

	return nil
}
