// Package metrics provides the service layer for metrics operations.
package metrics

// GetCounter retrieves the value of a counter metric by name.
func (s *metricsService) GetCounter(name string) (int64, error) {
	if s.dbRep != nil && s.dbRep.Ping() {
		return s.dbRep.ReadCounter(name)
	}

	return s.storageRep.ReadCounter(name)
}

// GetGauge retrieves the value of a gauge metric by name.
func (s *metricsService) GetGauge(name string) (float64, error) {
	if s.dbRep != nil && s.dbRep.Ping() {
		return s.dbRep.ReadGauge(name)
	}

	return s.storageRep.ReadGauge(name)
}

// GetMany retrieves all metrics as a map of name to string value.
func (s *metricsService) GetMany() map[string]string {
	if s.dbRep != nil && s.dbRep.Ping() {
		return s.dbRep.ReadAll()
	}

	return s.storageRep.ReadAll()
}
