package storage

import "fmt"

func (s *MemStorage) ReadCounter(name string) (int64, error) {
	value, exists := s.counter[name]
	if !exists {
		return 0, fmt.Errorf("counter metric '%s' not found", name)
	}
	return value, nil
}
