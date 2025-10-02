package main

import (
	"context"
	"fmt"
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
	cfg := config.NewAgentConfig()
	logger.Log.Info("agent cfg", zap.Any("cfg", cfg))

	a := agent.New(cfg)
	defer a.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	metricPoll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer metricPoll.Stop()

	metricReport := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer metricReport.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-metricReport.C:
			if err := a.SendMetricsBatch(); err != nil {
				fmt.Printf("something went wrong: %+v", err)
				return
			}
		case <-metricPoll.C:
			a.GetMetric()
		}
	}
}
