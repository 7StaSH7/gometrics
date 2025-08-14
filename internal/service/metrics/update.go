package metrics

import (
	"github.com/7StaSH7/gometrics/internal/model"
)

func (s *metricsService) UpdateMetric(mType, name string, value any) error {
	switch mType {
	case model.Gauge:
		s.storageRep.Replace(name, value.(float64))
	case model.Counter:
		s.storageRep.Add(name, value.(int64))
	}

	return nil
}
