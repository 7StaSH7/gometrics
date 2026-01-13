package config

import (
	"encoding/json"
	"os"
	"time"
)

// ServerJSONConfig holds server configuration from JSON file.
type ServerJSONConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

// AgentJSONConfig holds agent configuration from JSON file.
type AgentJSONConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// parseDuration parses a duration string and returns seconds.
func parseDuration(s string) int {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return int(d.Seconds())
}

// loadServerConfigFromFile loads server configuration from JSON file.
func loadServerConfigFromFile(path string) (*ServerJSONConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ServerJSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadAgentConfigFromFile loads agent configuration from JSON file.
func loadAgentConfigFromFile(path string) (*AgentJSONConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AgentJSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
