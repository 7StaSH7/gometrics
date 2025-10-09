package config

import (
	"flag"
	"log"

	"github.com/7StaSH7/gometrics/internal/config/db"
	"github.com/caarlos0/env"
)

type ServerConfig struct {
	LogLevel      string `env:"LOG_LEVEL"`
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	StoreFilePath string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	Key           string `env:"KEY"`
}

func NewServerConfig() (*ServerConfig, *db.PostgresConfig) {
	cfg := &ServerConfig{}
	psqlCfg := &db.PostgresConfig{}

	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to listen on")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "interval to store metrics to file")
	flag.StringVar(&cfg.StoreFilePath, "f", "metrics.json", "path to json file to store metrics")
	flag.BoolVar(&cfg.Restore, "r", false, "if need to restore from file first")
	flag.StringVar(&cfg.Key, "k", "", "key to calculate auth hash")

	flag.StringVar(&psqlCfg.URL, "d", "postgres://postgres:postgres@localhost:5432/metrics?search_path=public&sslmode=disable", "url for postgres db connection")

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		log.Panic(err)
	}
	if err := env.Parse(psqlCfg); err != nil {
		log.Panic(err)
	}

	return cfg, psqlCfg
}
