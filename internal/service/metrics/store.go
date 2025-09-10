package metrics

import (
	"context"
	"time"
)

func (s *metricsService) Store(ctx context.Context, restore bool, interval int) error {
	metricStore := time.NewTicker(time.Duration(interval) * time.Second)
	defer metricStore.Stop()

	if restore {
		if err := s.storageRep.Restore(); err != nil {
			panic(err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-metricStore.C:
			if err := s.storageRep.Store(); err != nil {
				panic(err)
			}
		}
	}
}
