package storage

import (
	"fmt"

	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
)

type MemStorage struct {
	gauges  map[string]float64
	counter map[string]int64
}

type MemStorageInterface interface {
	Replace(name string, value float64)
	Add(name string, value int64)
	ReadCounter(name string) int64
	ReadGauge(name string) float64
	ReadMany() map[string]string
}

func NewStorage() MemStorageInterface {
	return &MemStorage{
		gauges:  make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (s *MemStorage) Replace(name string, value float64) {
	logger.Log.Debug("replace value", zap.String("name", name), zap.Float64("value", value))
	s.gauges[name] = value
}

func (s *MemStorage) Add(name string, value int64) {
	logger.Log.Debug("add value", zap.String("name", name), zap.Int64("value", value))
	s.counter[name] += value
}

func (s *MemStorage) ReadCounter(name string) int64 {
	return s.counter[name]
}

func (s *MemStorage) ReadGauge(name string) float64 {
	return s.gauges[name]
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
