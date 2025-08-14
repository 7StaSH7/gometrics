package metrics

func (s *metricsService) GetOne(mType, name string) string {
	return s.storageRep.ReadOne(mType, name)
}
