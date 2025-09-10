package storage

import "fmt"

func (s *MemStorage) ReadGauge(name string) (float64, error) {
	value, exists := s.gauges[name]
	if !exists {
		return 0, fmt.Errorf("gauge metric '%s' not found", name)
	}
	return value, nil
}
