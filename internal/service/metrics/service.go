package metrics

import "github.com/7StaSH7/gometrics/internal/repository"

type MetricsService interface {
	UpdateMetric(mType, name string, value any) error
}

type metricsService struct {
	storageRep repository.MemStorageRepository
}

func New(storageRep repository.MemStorageRepository) MetricsService {
	return &metricsService{
		storageRep: storageRep,
	}
}
