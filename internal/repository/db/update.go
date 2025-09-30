package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (rep *databaseRepository) Add(tx pgx.Tx, name string, delta int64) error {
	sql := `
		insert into metrics (id, mType, delta) values ($1,'counter', $2)
		on conflict (id) do update
    set	delta = metrics.delta + excluded.delta;
  `
	if tx != nil {
		if _, err := tx.Exec(context.Background(), sql, name, delta); err != nil {
			return err
		}
	} else {
		if _, err := rep.db.Exec(context.Background(), sql, name, delta); err != nil {
			return err
		}
	}

	return nil
}

func (rep *databaseRepository) Replace(tx pgx.Tx, name string, value float64) error {
	sql := `
		insert into metrics (id, mType, value) values ($1, 'gauge', $2)
		on conflict (id) do update
    set	value = excluded.value;
  `
	if tx != nil {
		if _, err := tx.Exec(context.Background(), sql, name, value); err != nil {
			return err
		}
	} else {
		if _, err := rep.db.Exec(context.Background(), sql, name, value); err != nil {
			return err
		}
	}

	return nil
}
