package metrics

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/7StaSH7/gometrics/internal/repository/db"
	"github.com/7StaSH7/gometrics/internal/repository/storage"
	"github.com/jackc/pgx/v5"
)

type MetricsService interface {
	UpdateCounter(ctx context.Context, tx pgx.Tx, name string, value int64) error
	UpdateGauge(ctx context.Context, tx pgx.Tx, name string, value float64) error
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
	GetMany() map[string]string
	Store(ctx context.Context, restore bool, interval int) error
	Updates(ctx context.Context, metrics []model.Metrics) error
}

type metricsService struct {
	storageRep storage.MemStorageRepository
	dbRep      db.DatabaseRepository
}

func New(storageRep storage.MemStorageRepository, dbRep db.DatabaseRepository) MetricsService {
	return &metricsService{
		storageRep: storageRep,
		dbRep:      dbRep,
	}
}
