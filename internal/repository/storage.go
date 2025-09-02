package repository

import (
	"github.com/7StaSH7/gometrics/internal/storage"
)

type memStorageRepository struct {
	storage storage.MemStorageInterface
}

type MemStorageRepository interface {
	Replace(name string, value float64)
	Add(name string, value int64)
	ReadCounter(name string) (int64, error)
	ReadGauge(name string) (float64, error)
	ReadMany() map[string]string
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

func (s *memStorageRepository) ReadCounter(name string) (int64, error) {
	return s.storage.ReadCounter(name)
}

func (s *memStorageRepository) ReadGauge(name string) (float64, error) {
	return s.storage.ReadGauge(name)
}

func (s *memStorageRepository) ReadMany() map[string]string {
	return s.storage.ReadMany()
}
