package metrics

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/jackc/pgx/v5"
)

func (s *metricsService) UpdateCounter(ctx context.Context, tx pgx.Tx, name string, value int64) error {
	if s.dbRep.Ping() {
		if err := s.dbRep.Add(ctx, tx, name, value); err != nil {
			return err
		}
		return nil
	}

	if err := s.storageRep.Add(name, value); err != nil {
		return err
	}

	return nil
}

func (s *metricsService) UpdateGauge(ctx context.Context, tx pgx.Tx, name string, value float64) error {
	if s.dbRep.Ping() {
		if err := s.dbRep.Replace(ctx, tx, name, value); err != nil {
			return err
		}
		return nil
	}

	if err := s.storageRep.Replace(name, value); err != nil {
		return err
	}

	return nil
}

func (s *metricsService) Updates(ctx context.Context, metrics []model.Metrics) error {
	var tx pgx.Tx
	var err error

	if s.dbRep.Ping() {
		tx, err = s.dbRep.StartTransaction(ctx)
		if err != nil {
			return err
		}
	}

	for _, m := range metrics {
		switch m.MType {
		case model.Counter:
			err = s.UpdateCounter(ctx, tx, m.ID, *m.Delta)
		case model.Gauge:
			err = s.UpdateGauge(ctx, tx, m.ID, *m.Value)
		}
	}

	s.dbRep.IntrospectTransaction(ctx, tx, err)
	if err != nil {
		return err
	}

	return nil
}
