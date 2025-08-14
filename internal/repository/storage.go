package repository

import "github.com/7StaSH7/gometrics/internal/storage"

type memStorageRepository struct {
	storage storage.MemStorageInterface
}

type MemStorageRepository interface {
	Replace(name string, value float64)
	Add(name string, value int64)
}

func NewMemStorageRepository(storage storage.MemStorageInterface) MemStorageRepository {
	return &memStorageRepository{
		storage: storage,
	}
}

func (s *memStorageRepository) Replace(name string, value float64) {
	s.storage.Replace(name, value)
}

func (s *memStorageRepository) Add(name string, value int64) {
	s.storage.Add(name, value)
}
