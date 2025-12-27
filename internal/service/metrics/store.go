// Package metrics provides the service layer for metrics operations.
package metrics

import (
	"context"
	"time"
)

// Store periodically stores metrics to persistent storage.
func (s *metricsService) Store(ctx context.Context, restore bool, interval int) error {
	metricStore := time.NewTicker(time.Duration(interval) * time.Second)
	defer metricStore.Stop()

	if restore {
		if err := s.storageRep.Restore(); err != nil {
			return err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-metricStore.C:
			if err := s.storageRep.Store(); err != nil {
				return err
			}
		}
	}
}
