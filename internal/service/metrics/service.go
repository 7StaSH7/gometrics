package metrics

import "github.com/7StaSH7/gometrics/internal/repository"

type MetricsService interface {
	UpdateCounter(name string, value int64) error
	UpdateGauge(name string, value float64) error
	GetCounter(name string) int64
	GetGauge(name string) float64
	GetMany() map[string]string
}

type metricsService struct {
	storageRep repository.MemStorageRepository
}

func New(storageRep repository.MemStorageRepository) MetricsService {
	return &metricsService{
		storageRep: storageRep,
	}
}
