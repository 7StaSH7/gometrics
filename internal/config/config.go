package config

import (
	"flag"
	"log"
	"os"

	"github.com/7StaSH7/gometrics/internal/config/db"
	"github.com/caarlos0/env"
)

// ServerConfig holds configuration for the metrics server.
type ServerConfig struct {
	LogLevel      string `env:"LOG_LEVEL"`
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	StoreFilePath string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	Key           string `env:"KEY"`
	AuditFile     string `env:"AUDIT_FILE"`
	AuditURL      string `env:"AUDIT_URL"`
	CryptoKey     string `env:"CRYPTO_KEY"`
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
}

// NewServerConfig creates and parses the server configuration.
func NewServerConfig() (*ServerConfig, *db.PostgresConfig) {
	cfg := &ServerConfig{}
	psqlCfg := &db.PostgresConfig{}

	var configPath string
	flag.StringVar(&configPath, "c", "", "path to JSON config file")
	flag.StringVar(&configPath, "config", "", "path to JSON config file")

	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to listen on")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "interval to store metrics to file")
	flag.StringVar(&cfg.StoreFilePath, "f", "metrics.json", "path to json file to store metrics")
	flag.BoolVar(&cfg.Restore, "r", false, "if need to restore from file first")
	flag.StringVar(&cfg.Key, "k", "", "key to calculate auth hash")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "filepath to store audit events")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "url to send audit events")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "path to private key file for decryption")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet in CIDR notation")
	flag.StringVar(&cfg.TrustedSubnet, "trusted-subnet", "", "trusted subnet in CIDR notation")

	flag.StringVar(&psqlCfg.URL, "d", "postgres://postgres:postgres@localhost:5432/metrics?search_path=public&sslmode=disable", "url for postgres db connection")

	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		jsonCfg, err := loadServerConfigFromFile(configPath)
		if err != nil {
			log.Printf("failed to load config file: %v", err)
		} else {
			if jsonCfg.Address != "" {
				cfg.Address = jsonCfg.Address
			}
			cfg.Restore = jsonCfg.Restore
			if jsonCfg.StoreInterval != "" {
				cfg.StoreInterval = parseDuration(jsonCfg.StoreInterval)
			}
			if jsonCfg.StoreFile != "" {
				cfg.StoreFilePath = jsonCfg.StoreFile
			}
			if jsonCfg.DatabaseDSN != "" {
				psqlCfg.URL = jsonCfg.DatabaseDSN
			}
			if jsonCfg.CryptoKey != "" {
				cfg.CryptoKey = jsonCfg.CryptoKey
			}
			if jsonCfg.TrustedSubnet != "" {
				cfg.TrustedSubnet = jsonCfg.TrustedSubnet
			}
		}
	}

	if err := env.Parse(cfg); err != nil {
		log.Panic(err)
	}
	if err := env.Parse(psqlCfg); err != nil {
		log.Panic(err)
	}

	return cfg, psqlCfg
}
