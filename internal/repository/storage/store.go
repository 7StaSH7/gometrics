package storage

func (rep *memStorageRepository) Store() error {
	return rep.storage.Store()
}

func (rep *memStorageRepository) Restore() error {
	return rep.storage.Restore()
}
