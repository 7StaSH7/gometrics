package storage

func (s *memStorageRepository) Replace(name string, value float64) {
	s.storage.Replace(name, value)
}
