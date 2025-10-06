package pgerrors

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		return ClassifyPgError(pgErr)
	}

	return NonRetriable
}

func ClassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
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

func ExecuteWithRetry(ctx context.Context, pool *pgxpool.Pool, tx pgx.Tx, sql SQL) error {
	const maxRetries = 3
	var lastErr error

	classifier := NewPostgresErrorClassifier()

	var err error
	if pool != nil {
		_, err = pool.Exec(ctx, sql.Query, sql.Args...)
	} else if tx != nil {
		_, err = tx.Exec(ctx, sql.Query, sql.Args...)
	}
	if err == nil {
		return nil
	}

	lastErr = err
	classification := classifier.Classify(err)
	if classification == NonRetriable {
		return fmt.Errorf("error: %w", err)
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var delay time.Duration
		switch attempt {
		case 1:
			delay = 1 * time.Second
		case 2:
			delay = 3 * time.Second
		case 3:
			delay = 5 * time.Second
		}
		time.Sleep(delay)

		if pool != nil {
			_, err = pool.Exec(ctx, sql.Query, sql.Args...)
		} else if tx != nil {
			_, err = tx.Exec(ctx, sql.Query, sql.Args...)
		}
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("%d retries end with error: %w", maxRetries, lastErr)
}
