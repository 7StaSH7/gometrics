package storage

func (s *memStorageRepository) Store() error {
	return s.storage.Store()
}

func (s *memStorageRepository) Restore() error {
	return s.storage.Restore()
}
