package db

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (rep *databaseRepository) StartTransaction() (pgx.Tx, error) {
	return rep.db.Begin(context.Background())
}

func (rep *databaseRepository) IntrospectTransaction(tx pgx.Tx, err error) {
	ctx := context.Background()
	var e error
	if err != nil {
		logger.Log.Error("rollback", zap.Error(err))
		e = tx.Rollback(ctx)
	} else {
		e = tx.Commit(ctx)
	}

	if e != nil {
		logger.Log.Panic("ERROR", zap.Error(e))
	}
}
