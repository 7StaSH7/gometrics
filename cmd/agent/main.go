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
)

func main() {
	logger.Initialize("info")

	cfg := config.NewAgentConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a := agent.New(ctx, cfg)
	defer a.Close()

	metricPoll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer metricPoll.Stop()

	metricReport := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer metricReport.Stop()

	logger.Log.Info("agent started", zap.Any("config", cfg))

	for {
		select {
		case <-ctx.Done():
			return
		case <-metricReport.C:
			if err := a.SendMetricsBatch(); err != nil {
				logger.Log.Error("something went wrong", zap.Error(err))
				return
			}
		case <-metricPoll.C:
			a.GetMetric()
		}
	}
}
