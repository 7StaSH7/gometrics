package pgerrors

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return СlassifyPgError(pgErr)
	}

	return NonRetriable
}

func СlassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	case pgerrcode.TransactionRollback,
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected:
		return Retriable

	case pgerrcode.CannotConnectNow:
		return Retriable
	}

	switch pgErr.Code {
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	return NonRetriable
}

type SQL struct {
	Query string
	Args  []any
}

func ExecuteWithRetry(pool *pgxpool.Pool, tx pgx.Tx, sql SQL) error {
	const maxRetries = 3
	var lastErr error

	classifier := NewPostgresErrorClassifier()

	for _ = range maxRetries {
		var err error
		ctx := context.Background()
		if pool != nil {
			_, err = pool.Exec(ctx, sql.Query, sql.Args...)
			if err == nil {
				return nil
			}
		} else if tx != nil {
			_, err = tx.Exec(ctx, sql.Query, sql.Args...)
			if err == nil {
				return nil
			}
		}

		classification := classifier.Classify(err)
		if classification == NonRetriable {
			return fmt.Errorf("Error: %w\n", err)
		}
	}

	return fmt.Errorf("%d retries end with error: %w", maxRetries, lastErr)
}
