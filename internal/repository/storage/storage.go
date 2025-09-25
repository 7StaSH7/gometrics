package storage

import (
	"github.com/7StaSH7/gometrics/internal/storage"
)

type memStorageRepository struct {
	storage storage.MemStorageInterface
}

type MemStorageRepository interface {
	Replace(name string, value float64) error
	Add(name string, value int64) error
	ReadCounter(string) (int64, error)
	ReadGauge(string) (float64, error)
	ReadAll() map[string]string
	Restore() error
	Store() error
}

func NewMemStorageRepository(storage storage.MemStorageInterface) MemStorageRepository {
	return &memStorageRepository{
		storage: storage,
	}
}
