package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	Limit          int    `env:"RATE_LIMIT"`
}

func NewAgentConfig() *AgentConfig {
	cfg := &AgentConfig{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to send metrics to")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&cfg.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&cfg.Key, "k", "", "key to calculate auth hash")
	flag.IntVar(&cfg.Limit, "l", 5, "request rate limit")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		log.Panic(err)
	}

	return cfg
}
