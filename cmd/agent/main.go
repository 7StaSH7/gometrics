package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/7StaSH7/gometrics/internal/agent"
	"github.com/7StaSH7/gometrics/internal/config"
)

func main() {
	sCfg := config.NewServerConfig()
	aCfg := config.NewAgentConfig()
	flag.Parse()

	a := agent.New(sCfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	metricPoll := time.NewTicker(time.Duration(aCfg.PollInterval) * time.Second)
	defer metricPoll.Stop()

	metricReport := time.NewTicker(time.Duration(aCfg.ReportInterval) * time.Second)
	defer metricReport.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-metricReport.C:
			a.SendMetrics()
		case <-metricPoll.C:
			a.GetMetric()
		}
	}
}
