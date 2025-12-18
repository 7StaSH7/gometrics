package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	Up   = "up"
	Down = "down"
)

const (
	migrationDir = "file://migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func run() error {
	dir := parseArgs()

	_, cfg := config.NewServerConfig()

	m, err := migrate.New(migrationDir, cfg.URL)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	switch dir {
	case Up:
		if err := up(m); err != nil {
			return fmt.Errorf("migration up failed: %w", err)
		}
		fmt.Println("Migration up completed")
		return nil
	case Down:
		if err := down(m); err != nil {
			return fmt.Errorf("migration down failed: %w", err)
		}
		fmt.Println("Migration down completed")
		return nil
	default:
		return fmt.Errorf("invalid migration direction: %s", dir)
	}
}

func parseArgs() string {
	var dir string
	flag.StringVar(&dir, "dir", "up", "migration direction (up/down)")
	flag.Parse()

	return dir
}

func up(m *migrate.Migrate) error {
	if err := m.Up(); err != nil {
		return err
	}
	return nil
}

func down(m *migrate.Migrate) error {
	if err := m.Down(); err != nil {
		return err
	}
	return nil
}
