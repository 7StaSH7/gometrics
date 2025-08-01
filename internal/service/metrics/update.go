package metrics

import (
	"fmt"

	"github.com/7StaSH7/gometrics/internal/model"
)

func (s *metricsService) UpdateMetric(mType, name string, value any) error {
	switch mType {
	case model.Gauge:
		model.Storage.Replace(name, value.(float64))
	case model.Counter:
		model.Storage.Add(name, value.(int64))
	}

	fmt.Println(model.Storage)

	return nil
}
