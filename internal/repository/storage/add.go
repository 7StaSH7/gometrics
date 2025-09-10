package storage

func (s *memStorageRepository) Add(name string, value int64) {
	s.storage.Add(name, value)
}
