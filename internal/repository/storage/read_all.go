package storage

func (s *memStorageRepository) ReadAll() map[string]string {
	return s.storage.ReadAll()
}
