package db

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (rep *databaseRepository) StartTransaction(ctx context.Context) (pgx.Tx, error) {
	return rep.db.Begin(ctx)
}

func (rep *databaseRepository) IntrospectTransaction(ctx context.Context, tx pgx.Tx, err error) {
	if tx != nil {
		var e error
		if err != nil {
			e = tx.Rollback(ctx)
		} else {
			e = tx.Commit(ctx)
		}

		if e != nil {
			logger.Log.Panic("ERROR", zap.Error(e))
		}
	}
}
