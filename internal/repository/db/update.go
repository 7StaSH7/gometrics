package db

import "context"

func (rep *databaseRepository) Add(name string, value int64) error {
	if _, err := rep.db.Query(context.Background(), `
		insert into metrics (id, mType, delta) values ($1,'counter', $2)
		on conflict (id) do update
    set	delta = metrics.delta + excluded.delta;
    `, name, value); err != nil {
		return err
	}

	return nil
}

func (rep *databaseRepository) Replace(name string, value float64) error {
	if _, err := rep.db.Query(context.Background(), `
		insert into metrics (id, mType, value) values ($1, 'gauge', $2)
		on conflict (id) do update
    set	value = excluded.value;
    `, name, value); err != nil {
		return err
	}

	return nil
}
