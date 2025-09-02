package metrics

func (s *metricsService) UpdateCounter(name string, value int64) error {
	s.storageRep.Add(name, value)

	return nil
}

func (s *metricsService) UpdateGauge(name string, value float64) error {
	s.storageRep.Replace(name, value)

	return nil
}
