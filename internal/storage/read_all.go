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
