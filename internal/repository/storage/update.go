package storage

func (rep *memStorageRepository) Add(name string, value int64) error {
	rep.storage.Add(name, value)

	return nil
}

func (rep *memStorageRepository) Replace(name string, value float64) error {
	rep.storage.Replace(name, value)

	return nil
}
