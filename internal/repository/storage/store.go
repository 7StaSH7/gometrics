package storage

func (s *memStorageRepository) Store() error {
	return s.storage.Store()
}
