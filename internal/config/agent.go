package config

import (
	"flag"
)

type AgentConfig struct {
	PollInterval   int
	ReportInterval int
}

func NewAgentConfig() *AgentConfig {
	cfg := &AgentConfig{}

	flag.IntVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&cfg.PollInterval, "p", 2, "poll interval")
	flag.Parse()

	return cfg
}
