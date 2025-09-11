package db

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type PostgresConfig struct {
	URL string `env:"DATABASE_DSN"`
}

func NewPostgresConfig() *PostgresConfig {
	cfg := &PostgresConfig{}

	flag.StringVar(&cfg.URL, "d", "postgres://postgres:postgres@localhost:5432/metrics", "url for postgres db connection")

	if err := env.Parse(cfg); err != nil {
		log.Panic(err)
	}
	return cfg
}
