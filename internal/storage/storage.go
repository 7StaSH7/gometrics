package storage

type MemStorage struct {
	gauges  map[string]float64
	counter map[string]int64
}

type MemStorageInterface interface {
	Replace(name string, value float64)
	Add(name string, value int64)
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
