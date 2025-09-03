package metrics

func (s *metricsService) GetCounter(name string) (int64, error) {
	return s.storageRep.ReadCounter(name)
}

func (s *metricsService) GetGauge(name string) (float64, error) {
	return s.storageRep.ReadGauge(name)
}
