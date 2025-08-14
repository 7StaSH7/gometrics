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
	ReadOne(mType, name string) string
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

func (s *memStorageRepository) ReadOne(mType, name string) string {
	return s.storage.ReadOne(mType, name)
}

func (s *memStorageRepository) ReadMany() map[string]string {
	return s.storage.ReadMany()
}
