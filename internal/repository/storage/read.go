package storage

func (s *memStorageRepository) ReadAll() map[string]string {
	return s.storage.ReadAll()
}

func (s *memStorageRepository) ReadCounter(name string) (int64, error) {
	return s.storage.ReadCounter(name)
}

func (s *memStorageRepository) ReadGauge(name string) (float64, error) {
	return s.storage.ReadGauge(name)
}
