package db

import (
	"context"
	"fmt"

	"github.com/7StaSH7/gometrics/internal/model"
)

func (rep *databaseRepository) ReadAll() map[string]string {
	metrics := make(map[string]string, 0)
	rows, err := rep.db.Query(context.Background(), "select id, value, delta from metrics;")
	if err != nil {
		return map[string]string{}
	}
	defer rows.Close()

	for rows.Next() {
		var m model.Metrics
		err = rows.Scan(&m.ID, &m.Value, &m.Delta)
		if err != nil {
			return map[string]string{}
		}
		if m.Value != nil {
			metrics[m.ID] = fmt.Sprintf("%v", *m.Value)
		}
		if m.Delta != nil {
			metrics[m.ID] = fmt.Sprintf("%v", *m.Delta)
		}
	}

	return metrics
}

func (rep *databaseRepository) ReadCounter(name string) (int64, error) {
	var res int64

	if err := rep.db.QueryRow(context.Background(), "select delta from metrics where id = $1", name).Scan(&res); err != nil {
		return 0, err
	}

	return res, nil
}

func (rep *databaseRepository) ReadGauge(name string) (float64, error) {
	var res float64

	if err := rep.db.QueryRow(context.Background(), "select value from metrics where id = $1", name).Scan(&res); err != nil {
		return 0, err
	}

	return res, nil
}

func (rep *databaseRepository) Ping() bool {
	if rep.db == nil {
		return false
	}

	if err := rep.db.Ping(context.Background()); err != nil {
		return false
	}

	return true
}
