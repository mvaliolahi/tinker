package db

import (
	"fmt"
	"strings"

	"github.com/mvaliolahi/tinker/internal/config"
)

type Session struct {
	Driver string
	DSN    string
	Type   string
}

func NewSession(cfg *config.Database) (*Session, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no [database] section in tinker.toml")
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is empty (source: %s)", cfg.Source)
	}

	driver := cfg.Driver
	if driver == "" {
		driver = driverFromType(cfg.Type)
	}

	return &Session{Driver: driver, DSN: cfg.URL, Type: cfg.Type}, nil
}

func driverFromType(t string) string {
	switch strings.ToLower(t) {
	case "postgres", "postgresql":
		return "postgres"
	case "mysql":
		return "mysql"
	case "sqlite3", "sqlite":
		return "sqlite3"
	case "sqlserver", "mssql":
		return "sqlserver"
	case "mongodb", "mongo":
		return "mongo"
	default:
		return t
	}
}
