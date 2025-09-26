package metrics

func (s *metricsService) UpdateCounter(name string, value int64) error {
	if s.dbRep.Ping() {
		if err := s.dbRep.Add(name, value); err != nil {
			return err
		}
	}

	if err := s.storageRep.Add(name, value); err != nil {
		return err
	}

	return nil
}

func (s *metricsService) UpdateGauge(name string, value float64) error {
	if s.dbRep.Ping() {
		if err := s.dbRep.Replace(name, value); err != nil {
			return err
		}
	}

	if err := s.storageRep.Replace(name, value); err != nil {
		return err
	}

	return nil
}
