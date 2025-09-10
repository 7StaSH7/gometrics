package storage

func (s *memStorageRepository) ReadCounter(name string) (int64, error) {
	return s.storage.ReadCounter(name)
}
