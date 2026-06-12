package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a database migration file.
type Migration struct {
	Version  string // e.g., "001"
	Name     string // e.g., "create_users"
	UpFile   string // full path to .up.sql
	DownFile string // full path to .down.sql, may be empty
}

// MigrateUp runs all pending migrations (not yet applied) in order.
// It uses a version tracking table `_tinker_migrations` to track applied versions.
func (s *Session) MigrateUp(dir string) ([]string, error) {
	migrations, err := parseMigrations(dir)
	if err != nil {
		return nil, err
	}
	if len(migrations) == 0 {
		return nil, fmt.Errorf("no migration files found in %s", dir)
	}

	if err := s.ensureMigrationTable(); err != nil {
		return nil, fmt.Errorf("creating migration table: %w", err)
	}

	applied, err := s.appliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("checking applied migrations: %w", err)
	}

	var pending []Migration
	for _, m := range migrations {
		if !applied[m.Version] {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		return nil, nil
	}

	var appliedNames []string
	for _, m := range pending {
		if err := s.applyMigration(m); err != nil {
			return appliedNames, fmt.Errorf("migration %s: %w", m.Version+"_"+m.Name, err)
		}
		appliedNames = append(appliedNames, m.Version+"_"+m.Name)
	}

	return appliedNames, nil
}

// MigrateDown rolls back the most recent migration.
func (s *Session) MigrateDown(dir string) (string, error) {
	migrations, err := parseMigrations(dir)
	if err != nil {
		return "", err
	}

	if err := s.ensureMigrationTable(); err != nil {
		return "", fmt.Errorf("creating migration table: %w", err)
	}

	applied, err := s.appliedMigrations()
	if err != nil {
		return "", fmt.Errorf("checking applied migrations: %w", err)
	}

	// Find the last applied migration
	var last *Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if applied[migrations[i].Version] {
			last = &migrations[i]
			break
		}
	}

	if last == nil {
		return "", nil // nothing to roll back
	}

	if last.DownFile == "" {
		return "", fmt.Errorf("no down migration for %s_%s", last.Version, last.Name)
	}

	if err := s.rollbackMigration(*last); err != nil {
		return "", fmt.Errorf("rollback %s: %w", last.Version+"_"+last.Name, err)
	}

	return last.Version + "_" + last.Name, nil
}

// MigrateStatus returns a list of all migrations with their applied status.
func (s *Session) MigrateStatus(dir string) ([]MigrationStatus, error) {
	migrations, err := parseMigrations(dir)
	if err != nil {
		return nil, err
	}

	if err := s.ensureMigrationTable(); err != nil {
		return nil, err
	}

	applied, err := s.appliedMigrations()
	if err != nil {
		return nil, err
	}

	var status []MigrationStatus
	for _, m := range migrations {
		status = append(status, MigrationStatus{
			Migration: m,
			Applied:   applied[m.Version],
		})
	}
	return status, nil
}

// MigrationStatus is a migration with its applied state.
type MigrationStatus struct {
	Migration
	Applied bool
}

// ensureMigrationTable creates the _tinker_migrations table if it doesn't exist.
func (s *Session) ensureMigrationTable() error {
	var q string
	switch s.Type {
	case "sqlite3":
		q = `CREATE TABLE IF NOT EXISTS _tinker_migrations (version TEXT PRIMARY KEY, name TEXT NOT NULL, applied_at DATETIME DEFAULT CURRENT_TIMESTAMP)`
	case "postgres":
		q = `CREATE TABLE IF NOT EXISTS _tinker_migrations (version TEXT PRIMARY KEY, name TEXT NOT NULL, applied_at TIMESTAMP DEFAULT NOW())`
	case "mysql":
		q = `CREATE TABLE IF NOT EXISTS _tinker_migrations (version VARCHAR(255) PRIMARY KEY, name VARCHAR(255) NOT NULL, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`
	default:
		q = `CREATE TABLE IF NOT EXISTS _tinker_migrations (version TEXT PRIMARY KEY, name TEXT NOT NULL, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`
	}
	return s.execSQL(q)
}

// appliedMigrations returns a set of already-applied migration versions.
func (s *Session) appliedMigrations() (map[string]bool, error) {
	var rows func() (map[string]bool, error)
	if s.db != nil {
		rows = func() (map[string]bool, error) {
			result := make(map[string]bool)
			rs, err := s.db.Query("SELECT version FROM _tinker_migrations ORDER BY version")
			if err != nil {
				return nil, err
			}
			defer rs.Close()
			for rs.Next() {
				var v string
				if err := rs.Scan(&v); err != nil {
					return nil, err
				}
				result[v] = true
			}
			return result, rs.Err()
		}
	} else {
		rows = func() (map[string]bool, error) {
			out, err := s.Exec("SELECT version FROM _tinker_migrations ORDER BY version;")
			if err != nil {
				return nil, err
			}
			result := make(map[string]bool)
			for _, line := range strings.Split(out, "\n") {
				v := strings.TrimSpace(line)
				if v != "" && v != "version" && !strings.HasPrefix(v, "-") {
					result[v] = true
				}
			}
			return result, nil
		}
	}
	return rows()
}

// applyMigration runs the up migration and records it.
func (s *Session) applyMigration(m Migration) error {
	data, err := os.ReadFile(m.UpFile)
	if err != nil {
		return fmt.Errorf("reading up file: %w", err)
	}

	stmts := splitSQL(string(data))
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := s.execSQL(stmt); err != nil {
			return fmt.Errorf("executing up: %w", err)
		}
	}

	// Record the migration
	return s.execSQL(fmt.Sprintf("INSERT INTO _tinker_migrations (version, name) VALUES ('%s', '%s')",
		escapeSingle(m.Version), escapeSingle(m.Name)))
}

// rollbackMigration runs the down migration and removes the record.
func (s *Session) rollbackMigration(m Migration) error {
	data, err := os.ReadFile(m.DownFile)
	if err != nil {
		return fmt.Errorf("reading down file: %w", err)
	}

	stmts := splitSQL(string(data))
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := s.execSQL(stmt); err != nil {
			return fmt.Errorf("executing down: %w", err)
		}
	}

	// Remove the migration record
	return s.execSQL(fmt.Sprintf("DELETE FROM _tinker_migrations WHERE version = '%s'",
		escapeSingle(m.Version)))
}

// parseMigrations reads migration files from a directory.
// Expected format: 001_create_users.up.sql / 001_create_users.down.sql
func parseMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	// Map version -> Migration
	migrations := make(map[string]*Migration)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			continue
		}

		name := e.Name()
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		rest := parts[1]

		var isUp, isDown bool
		switch {
		case strings.HasSuffix(rest, ".up.sql"):
			isUp = true
			rest = strings.TrimSuffix(rest, ".up.sql")
		case strings.HasSuffix(rest, ".down.sql"):
			isDown = true
			rest = strings.TrimSuffix(rest, ".down.sql")
		default:
			continue
		}

		if _, ok := migrations[version]; !ok {
			migrations[version] = &Migration{Version: version, Name: rest}
		}
		migrations[version].Name = rest // normalize

		fullPath := filepath.Join(dir, name)
		if isUp {
			migrations[version].UpFile = fullPath
		} else if isDown {
			migrations[version].DownFile = fullPath
		}
	}

	// Sort by version
	var versions []string
	for v := range migrations {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	var result []Migration
	for _, v := range versions {
		m := migrations[v]
		if m.UpFile != "" { // Must have at least an up file
			result = append(result, *m)
		}
	}

	return result, nil
}
