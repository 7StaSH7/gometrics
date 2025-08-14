package main

import (
	"time"

	"github.com/7StaSH7/gometrics/internal/agent"
)

var pollInterval, reportInterval = 2 * time.Second, 10 * time.Second

func main() {
	a := agent.New()

	mt := time.NewTicker(pollInterval)
	go func() {
		for range mt.C {
			a.GetMetric()
		}
	}()

	for {
		a.SendMetrics()
		time.Sleep(reportInterval)
	}
}
