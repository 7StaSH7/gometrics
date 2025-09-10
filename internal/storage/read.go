package storage

import "fmt"

func (s *MemStorage) ReadAll() map[string]string {
	result := make(map[string]string)

	for name, value := range s.counter {
		result[name] = fmt.Sprint(value)
	}

	for name, value := range s.gauges {
		result[name] = fmt.Sprint(value)
	}

	return result
}

func (s *MemStorage) ReadCounter(name string) (int64, error) {
	value, exists := s.counter[name]
	if !exists {
		return 0, fmt.Errorf("counter metric '%s' not found", name)
	}
	return value, nil
}

func (s *MemStorage) ReadGauge(name string) (float64, error) {
	value, exists := s.gauges[name]
	if !exists {
		return 0, fmt.Errorf("gauge metric '%s' not found", name)
	}
	return value, nil
}
