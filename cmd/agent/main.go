package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/7StaSH7/gometrics/internal/agent"
	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger.Initialize("info")

	cfg := config.NewAgentConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	a := agent.New(gCtx, cfg)
	defer a.Close()

	logger.Log.Info("agent started", zap.Any("config", cfg))

	sendJobs := make(chan func() error, cfg.Limit)

	g.Go(func() error {
		t := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer t.Stop()

		for {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case <-t.C:
				if err := a.GetRuntimeMetrics(); err != nil {
					logger.Log.Error("get runtime metrics error", zap.Error(err))
				}
			}
		}
	})

	g.Go(func() error {
		t := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer t.Stop()

		for {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case <-t.C:
				if err := a.GetGopsutilMetrics(); err != nil {
					logger.Log.Error("get gopsutil metrics error", zap.Error(err))
				}
			}
		}
	})

	g.Go(func() error {
		ticker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case <-ticker.C:
				sendJobs <- func() error {
					return a.SendMetricsBatch()
				}
			}
		}
	})

	for w := 1; w <= cfg.Limit; w++ {
		g.Go(func() error {
			worker(gCtx, w, sendJobs)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.Log.Error("something went wrong", zap.Error(err))
		panic(err)
	}
}

func worker(ctx context.Context, id int, jobs <-chan func() error) {
	logger.Log.Info("worker started", zap.Int("id", id))
	for {
		select {
		case j, ok := <-jobs:
			if !ok {
				return
			}
			if err := j(); err != nil {
				logger.Log.Error("error in job", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}
