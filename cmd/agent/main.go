package main

import (
	"flag"
	"time"

	"github.com/7StaSH7/gometrics/internal/agent"
)

var args struct {
	a string
	r int
	p int
}

func main() {
	a := agent.New(args.a)

	mt := time.NewTicker(time.Duration(args.p) * time.Second)
	mr := time.NewTicker(time.Duration(args.r) * time.Second)

	for {
		select {
		case <-mr.C:
			a.SendMetrics()
		case <-mt.C:
			a.GetMetric()
		}
	}
}

func init() {
	flag.StringVar(&args.a, "a", "localhost:8080", "address to listen on")
	flag.IntVar(&args.r, "r", 10, "report interval")
	flag.IntVar(&args.p, "p", 2, "poll interval")
	flag.Parse()
}
