// Package config provides configuration structures and parsing for the metrics application.
package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
)

// AgentConfig holds configuration for the metrics agent.
type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	Limit          int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
}

// NewAgentConfig creates and parses the agent configuration.
func NewAgentConfig() *AgentConfig {
	cfg := &AgentConfig{}

	var configPath string
	flag.StringVar(&configPath, "c", "", "path to JSON config file")
	flag.StringVar(&configPath, "config", "", "path to JSON config file")

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to send metrics to")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&cfg.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&cfg.Key, "k", "", "key to calculate auth hash")
	flag.IntVar(&cfg.Limit, "l", 5, "request rate limit")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "path to public key file for encryption")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		jsonCfg, err := loadAgentConfigFromFile(configPath)
		if err != nil {
			log.Printf("failed to load config file: %v", err)
		} else {
			if jsonCfg.Address != "" {
				cfg.Address = jsonCfg.Address
			}
			if jsonCfg.ReportInterval != "" {
				cfg.ReportInterval = parseDuration(jsonCfg.ReportInterval)
			}
			if jsonCfg.PollInterval != "" {
				cfg.PollInterval = parseDuration(jsonCfg.PollInterval)
			}
			if jsonCfg.CryptoKey != "" {
				cfg.CryptoKey = jsonCfg.CryptoKey
			}
		}
	}

	if err := env.Parse(cfg); err != nil {
		log.Panic(err)
	}

	return cfg
}
