package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type ServerConfig struct {
	LogLevel      string `env:"LOG_LEVEL"`
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	StoreFilePath string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}

	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to listen on")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "interval to store metrics to file")
	flag.StringVar(&cfg.StoreFilePath, "f", "metrics.json", "path to json file to store metrics")
	flag.BoolVar(&cfg.Restore, "r", false, "if need to restore from file first")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		log.Panic(err)
	}

	return cfg
}
