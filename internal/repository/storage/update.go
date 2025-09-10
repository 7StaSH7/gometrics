package storage

func (s *memStorageRepository) Add(name string, value int64) {
	s.storage.Add(name, value)
}

func (s *memStorageRepository) Replace(name string, value float64) {
	s.storage.Replace(name, value)
}
