package storage

import (
	"fmt"

	"github.com/7StaSH7/gometrics/internal/model"
)

type MemStorage struct {
	gauges  map[string]float64
	counter map[string]int64
}

type MemStorageInterface interface {
	Replace(name string, value float64)
	Add(name string, value int64)
	ReadOne(mType, name string) string
	ReadMany() map[string]string
}

func NewStorage() MemStorageInterface {
	return &MemStorage{
		gauges:  make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (s *MemStorage) Replace(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemStorage) Add(name string, value int64) {
	s.counter[name] += value
}

func (s *MemStorage) ReadOne(mType, name string) string {
	switch mType {
	case model.Counter:
		v, ok := s.counter[name]
		if !ok {
			return ""
		}

		return fmt.Sprint(v)
	case model.Gauge:
		v, ok := s.gauges[name]
		if !ok {
			return ""
		}

		return fmt.Sprint(v)
	}

	return ""
}

func (s *MemStorage) ReadMany() map[string]string {
	result := make(map[string]string)

	for name, value := range s.counter {
		result[name] = fmt.Sprint(value)
	}

	for name, value := range s.gauges {
		result[name] = fmt.Sprint(value)
	}

	return result
}
