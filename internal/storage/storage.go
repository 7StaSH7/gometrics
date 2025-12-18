package storage

import (
	"github.com/7StaSH7/gometrics/internal/config"
)

// MemStorage represents an in-memory storage for metrics.
type MemStorage struct {
	gauges   map[string]float64
	counter  map[string]int64
	filePath string
	isSync   bool
}

// MemStorageInterface defines the interface for memory storage operations.
type MemStorageInterface interface {
	Replace(name string, value float64)
	Add(name string, value int64)
	ReadCounter(name string) (int64, error)
	ReadGauge(name string) (float64, error)
	ReadAll() map[string]string
	Store() error
	Restore() error
}

// NewStorage creates a new MemStorage instance with the given server config.
func NewStorage(cfg *config.ServerConfig) MemStorageInterface {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counter:  make(map[string]int64),
		filePath: cfg.StoreFilePath,
		isSync:   cfg.StoreInterval == 0,
	}
}
