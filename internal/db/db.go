// Package db provides database interaction capabilities by wrapping usql.
// It resolves connection details from tinker.toml and opens an interactive
// database session using usql as the underlying tool.
package db

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mvaliolahi/tinker/internal/config"
)

// Session represents a database session configuration.
type Session struct {
	Driver string
	DSN    string
	Type   string
}

// NewSession creates a new database session from config.
func NewSession(cfg *config.DatabaseConfig) (*Session, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no database configuration found in tinker.toml")
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is empty after resolving source: %s", cfg.Source)
	}

	driver := cfg.Driver
	if driver == "" {
		driver = driverFromType(cfg.Type)
	}

	return &Session{
		Driver: driver,
		DSN:    cfg.URL,
		Type:   cfg.Type,
	}, nil
}

// Connect opens an interactive usql session with the configured database.
func (s *Session) Connect() error {
	usqlPath, err := exec.LookPath("usql")
	if err != nil {
		return fmt.Errorf("usql is not installed. Install it with:\n  go install github.com/xo/usql@latest\n  or: brew install usql")
	}

	dsn := s.DSN

	// If the DSN doesn't already have a driver prefix, add one for usql
	if !strings.Contains(dsn, "://") && !strings.HasPrefix(dsn, "/") {
		// usql expects driver:// or driver: format
		dsn = fmt.Sprintf("%s://%s", s.Driver, dsn)
	}

	cmd := exec.Command(usqlPath, dsn)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Exec runs a single SQL command against the database and returns output.
func (s *Session) Exec(query string) (string, error) {
	usqlPath, err := exec.LookPath("usql")
	if err != nil {
		return "", fmt.Errorf("usql is not installed")
	}

	dsn := s.DSN
	if !strings.Contains(dsn, "://") && !strings.HasPrefix(dsn, "/") {
		dsn = fmt.Sprintf("%s://%s", s.Driver, dsn)
	}

	cmd := exec.Command(usqlPath, dsn, "-c", query)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Describe returns the schema description for a table.
func (s *Session) Describe(tableName string) (string, error) {
	var query string
	switch s.Type {
	case "postgres":
		query = fmt.Sprintf("SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position;", tableName)
	case "mysql":
		query = fmt.Sprintf("DESCRIBE %s;", tableName)
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	default:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 0;", tableName)
	}
	return s.Exec(query)
}

// Tables lists all tables in the database.
func (s *Session) Tables() (string, error) {
	var query string
	switch s.Type {
	case "postgres":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;"
	case "mysql":
		query = "SHOW TABLES;"
	case "sqlite3":
		query = "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;"
	default:
		query = "SELECT table_name FROM information_schema.tables ORDER BY table_name;"
	}
	return s.Exec(query)
}

// Count returns the row count for a given table with optional WHERE clause.
func (s *Session) Count(table string, where string) (string, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if where != "" {
		query += fmt.Sprintf(" WHERE %s", where)
	}
	query += ";"
	return s.Exec(query)
}

// Find returns a single row by ID.
func (s *Session) Find(table string, id string) (string, error) {
	return s.Exec(fmt.Sprintf("SELECT * FROM %s WHERE id = %s LIMIT 1;", table, id))
}

// driverFromType maps database type names to usql driver names.
func driverFromType(dbType string) string {
	switch strings.ToLower(dbType) {
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
		return dbType
	}
}
