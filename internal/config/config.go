package config

import "flag"

type ServerConfig struct {
	Address string
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to listen on")
	flag.Parse()

	return cfg
}
