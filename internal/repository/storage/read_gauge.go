package storage

func (s *memStorageRepository) ReadGauge(name string) (float64, error) {
	return s.storage.ReadGauge(name)
}
