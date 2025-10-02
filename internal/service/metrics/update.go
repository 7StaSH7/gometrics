package metrics

import (
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/jackc/pgx/v5"
)

func (s *metricsService) UpdateCounter(tx pgx.Tx, name string, value int64) error {
	if s.dbRep.Ping() {
		if err := s.dbRep.Add(tx, name, value); err != nil {
			return err
		}
	}

	if err := s.storageRep.Add(name, value); err != nil {
		return err
	}

	return nil
}

func (s *metricsService) UpdateGauge(tx pgx.Tx, name string, value float64) error {
	if s.dbRep.Ping() {
		if err := s.dbRep.Replace(tx, name, value); err != nil {
			return err
		}
	}

	if err := s.storageRep.Replace(name, value); err != nil {
		return err
	}

	return nil
}

func (s *metricsService) Updates(metrics []model.Metrics) error {
	var tx pgx.Tx
	var err error

	if s.dbRep.Ping() {
		tx, err = s.dbRep.StartTransaction()
		defer s.dbRep.IntrospectTransaction(tx, err)
	}

	for _, m := range metrics {
		switch m.MType {
		case model.Counter:
			err = s.UpdateCounter(tx, m.ID, *m.Delta)
		case model.Gauge:
			err = s.UpdateGauge(tx, m.ID, *m.Value)
		}
	}

	if err != nil {
		return err
	}

	return nil
}
