package storage

func (rep *memStorageRepository) ReadAll() map[string]string {
	return rep.storage.ReadAll()
}

func (rep *memStorageRepository) ReadCounter(name string) (int64, error) {
	return rep.storage.ReadCounter(name)
}

func (rep *memStorageRepository) ReadGauge(name string) (float64, error) {
	return rep.storage.ReadGauge(name)
}
