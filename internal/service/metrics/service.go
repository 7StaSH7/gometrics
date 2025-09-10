package metrics

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/repository/storage"
)

type MetricsService interface {
	UpdateCounter(name string, value int64) error
	UpdateGauge(name string, value float64) error
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
	GetMany() map[string]string
	Store(ctx context.Context, restore bool, interval int) error
}

type metricsService struct {
	storageRep storage.MemStorageRepository
}

func New(storageRep storage.MemStorageRepository) MetricsService {
	return &metricsService{
		storageRep: storageRep,
	}
}
