package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	UP   = "up"
	DOWN = "down"
)

const (
	MIGRATION_DIR = "file://migrations"
)

func main() {
	dir := parseArgs()

	_, cfg := config.NewServerConfig()

	fmt.Println(dir, cfg.URL)

	m, err := migrate.New(MIGRATION_DIR, cfg.URL)
	if err != nil {
		log.Fatal(err)
	}

	switch dir {
	case UP:
		{
			if err := up(m); err != nil {
				log.Fatal(err)
			}
			log.Println("Migration up completed")
			os.Exit(0)
		}
	case DOWN:
		{
			if err := down(m); err != nil {
				log.Fatal(err)
			}
			log.Println("Migration down completed")
			os.Exit(0)
		}
	default:
		{
			log.Fatal("Invalid migration direction")
		}
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
