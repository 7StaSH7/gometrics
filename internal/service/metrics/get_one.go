package metrics

func (s *metricsService) GetCounter(name string) int64 {
	return s.storageRep.ReadCounter(name)
}

func (s *metricsService) GetGauge(name string) float64 {
	return s.storageRep.ReadGauge(name)
}