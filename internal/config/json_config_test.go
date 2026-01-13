package config

import (
	"os"
	"testing"
)

func TestLoadServerConfigFromFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "server_config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	jsonContent := `{
		"address": "localhost:9090",
		"restore": true,
		"store_interval": "5s",
		"store_file": "/tmp/metrics.db",
		"database_dsn": "postgres://user:pass@localhost:5432/db",
		"crypto_key": "/tmp/key.pem"
	}`

	if _, err := tmpfile.Write([]byte(jsonContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadServerConfigFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Errorf("expected address localhost:9090, got %s", cfg.Address)
	}
	if !cfg.Restore {
		t.Error("expected restore to be true")
	}
	if cfg.StoreInterval != "5s" {
		t.Errorf("expected store_interval 5s, got %s", cfg.StoreInterval)
	}
	if cfg.StoreFile != "/tmp/metrics.db" {
		t.Errorf("expected store_file /tmp/metrics.db, got %s", cfg.StoreFile)
	}
	if cfg.DatabaseDSN != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("expected database_dsn postgres://user:pass@localhost:5432/db, got %s", cfg.DatabaseDSN)
	}
	if cfg.CryptoKey != "/tmp/key.pem" {
		t.Errorf("expected crypto_key /tmp/key.pem, got %s", cfg.CryptoKey)
	}
}

func TestLoadAgentConfigFromFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "agent_config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	jsonContent := `{
		"address": "localhost:9090",
		"report_interval": "10s",
		"poll_interval": "2s",
		"crypto_key": "/tmp/pubkey.pem"
	}`

	if _, err := tmpfile.Write([]byte(jsonContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadAgentConfigFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Errorf("expected address localhost:9090, got %s", cfg.Address)
	}
	if cfg.ReportInterval != "10s" {
		t.Errorf("expected report_interval 10s, got %s", cfg.ReportInterval)
	}
	if cfg.PollInterval != "2s" {
		t.Errorf("expected poll_interval 2s, got %s", cfg.PollInterval)
	}
	if cfg.CryptoKey != "/tmp/pubkey.pem" {
		t.Errorf("expected crypto_key /tmp/pubkey.pem, got %s", cfg.CryptoKey)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"seconds", "10s", 10},
		{"minutes", "2m", 120},
		{"hours", "1h", 3600},
		{"empty", "", 0},
		{"invalid", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDuration(tt.input)
			if got != tt.expected {
				t.Errorf("parseDuration(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}
