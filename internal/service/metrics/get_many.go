package metrics

func (s *metricsService) GetMany() map[string]string {
	return s.storageRep.ReadMany()
}
