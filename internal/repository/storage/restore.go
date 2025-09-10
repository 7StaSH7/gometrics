package storage

func (s *memStorageRepository) Restore() error {
	return s.storage.Restore()
}
