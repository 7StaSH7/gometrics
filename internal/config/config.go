package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type ServerConfig struct {
	LogLevel string `env:"LOG_LEVEL"`
	Address  string `env:"ADDRESS"`
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}

	flag.StringVar(&cfg.LogLevel, "l", "debug", "log level")
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to listen on")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		log.Panic(err)
	}

	return cfg
}
