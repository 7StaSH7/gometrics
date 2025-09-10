package storage

import (
	"github.com/7StaSH7/gometrics/internal/model"
)

func (s *MemStorage) Store() error {
	metrics := make([]model.Metrics, 0)
	for name, value := range s.gauges {
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Gauge,
			Value: &value,
		})

	}
	for name, value := range s.counter {
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Counter,
			Delta: &value,
		})
	}
	
	s.write(metrics)

	return nil
}
