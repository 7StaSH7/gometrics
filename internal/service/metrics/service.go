package metrics

import "github.com/7StaSH7/gometrics/internal/repository"

type MetricsService interface {
	Update(mType, name string, value any) error
	GetOne(mType, name string) string
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
