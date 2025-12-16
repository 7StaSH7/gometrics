package storage

func (s *MemStorage) Replace(name string, value float64) {
	s.gauges[name] = value
	if s.isSync {
		s.Store()
	}
}

func (s *MemStorage) Add(name string, value int64) {
	s.counter[name] += value
	if s.isSync {
		s.Store()
	}
}
