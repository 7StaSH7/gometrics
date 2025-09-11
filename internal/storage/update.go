package storage

import (
	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
)

func (s *MemStorage) Replace(name string, value float64) {
	logger.Log.Debug("replace value", zap.String("name", name), zap.Float64("value", value))
	s.gauges[name] = value
	if s.isSync {
		s.Store()
	}
}

func (s *MemStorage) Add(name string, value int64) {
	logger.Log.Debug("add value", zap.String("name", name), zap.Int64("value", value))
	s.counter[name] += value
	if s.isSync {
		s.Store()
	}
}
