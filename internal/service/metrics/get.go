package metrics

func (s *metricsService) GetCounter(name string) (int64, error) {
	if s.dbRep.Ping() {
		return s.dbRep.ReadCounter(name)
	}

	return s.storageRep.ReadCounter(name)
}

func (s *metricsService) GetGauge(name string) (float64, error) {
	if s.dbRep.Ping() {
		return s.dbRep.ReadGauge(name)
	}

	return s.storageRep.ReadGauge(name)
}

func (s *metricsService) GetMany() map[string]string {
	if s.dbRep.Ping() {
		return s.dbRep.ReadAll()
	}

	return s.storageRep.ReadAll()
}
