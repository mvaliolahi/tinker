package db

import (
	"fmt"
	"strings"
)

// Ping tests database connectivity. Returns nil on success.
func (s *Session) Ping() error {
	if s.db != nil {
		return s.db.Ping()
	}
	// No native connection — try a trivial CLI query
	_, err := s.Exec("SELECT 1;")
	return err
}

// Size returns row counts for all tables.
func (s *Session) Size() (string, error) {
	if s.db != nil {
		return s.sizeNative()
	}
	return s.sizeCLI()
}

func (s *Session) sizeNative() (string, error) {
	switch s.Type {
	case "sqlite3":
		// Get all table names and count each one
		names, err := s.tableNamesNative()
		if err != nil {
			return "", err
		}
		return s.countTablesRendered(names)
	case "postgres":
		rows, err := s.db.Query(
			"SELECT relname as \"Table\", n_live_tup as \"Rows\" FROM pg_stat_user_tables ORDER BY n_live_tup DESC;",
		)
		if err != nil {
			return "", fmt.Errorf("size query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	case "mysql":
		rows, err := s.db.Query(
			"SELECT table_name as \"Table\", table_rows as \"Rows\" FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_rows DESC;",
		)
		if err != nil {
			return "", fmt.Errorf("size query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	default:
		return s.sizeCLI()
	}
}

// tableNamesNative returns just the table names as a slice (no formatting).
func (s *Session) tableNamesNative() ([]string, error) {
	var q string
	switch s.Type {
	case "postgres":
		q = "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename"
	case "mysql":
		q = "SHOW TABLES"
	case "sqlite3":
		q = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	default:
		q = "SELECT table_name FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() ORDER BY table_name"
	}

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying tables: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// countTablesRendered counts rows in each table and renders a go-pretty table.
func (s *Session) countTablesRendered(tableNames []string) (string, error) {
	headers := []string{"Table", "Rows"}
	var dataRows [][]string

	for _, t := range tableNames {
		var count int64
		q := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(t))
		if err := s.db.QueryRow(q).Scan(&count); err != nil {
			dataRows = append(dataRows, []string{t, "?"})
			continue
		}
		dataRows = append(dataRows, []string{t, fmt.Sprintf("%d", count)})
	}

	return renderTable(headers, dataRows), nil
}

func (s *Session) sizeCLI() (string, error) {
	switch s.Type {
	case "sqlite3":
		return s.sizeCLISQLite()
	case "postgres":
		q := "SELECT relname as \"Table\", n_live_tup as \"Rows\" FROM pg_stat_user_tables ORDER BY n_live_tup DESC;"
		return s.ExecFormatted(q)
	case "mysql":
		q := "SELECT table_name as \"Table\", table_rows as \"Rows\" FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_rows DESC;"
		return s.ExecFormatted(q)
	default:
		q := "SELECT table_name, table_rows FROM information_schema.tables ORDER BY table_rows DESC;"
		return s.ExecFormatted(q)
	}
}

// sizeCLISQLite counts rows per table using the sqlite3 CLI.
// It fetches the table list, then runs SELECT COUNT(*) for each one.
func (s *Session) sizeCLISQLite() (string, error) {
	// Get table names via the tablesCLI path
	tablesOut, err := s.tablesCLI()
	if err != nil {
		return "", err
	}

	// Parse table names (one per line, trimmed)
	var names []string
	for _, line := range strings.Split(tablesOut, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			names = append(names, t)
		}
	}
	if len(names) == 0 {
		return "", nil
	}

	// Count rows in each table
	headers := []string{"Table", "Rows"}
	var dataRows [][]string
	for _, t := range names {
		q := fmt.Sprintf("SELECT COUNT(*) FROM %s;", quoteIdent(t))
		out, err := s.Exec(q)
		if err != nil {
			dataRows = append(dataRows, []string{t, "?"})
			continue
		}
		count := strings.TrimSpace(out)
		dataRows = append(dataRows, []string{t, count})
	}

	return renderTable(headers, dataRows), nil
}

// Count returns the count of rows in a table. Table name is validated;
// WHERE clause is treated as a raw SQL condition (user is responsible for safety).
func (s *Session) Count(table, where string) (string, error) {
	if err := s.validateTable(table); err != nil {
		return "", err
	}

	if s.db != nil {
		var count int64
		var q string
		if where != "" {
			q = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", quoteIdent(table), where)
		} else {
			q = fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(table))
		}
		err := s.db.QueryRow(q).Scan(&count)
		if err != nil {
			return "", fmt.Errorf("count query: %w", err)
		}
		return fmt.Sprintf("%d\n", count), nil
	}

	// Fallback to CLI (table already validated above)
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(table))
	if where != "" {
		q += fmt.Sprintf(" WHERE %s", where)
	}
	return s.Exec(q + ";")
}

// Find returns a single row by ID. Table name is validated; ID uses parameterized query.
func (s *Session) Find(table, id string) (string, error) {
	if err := s.validateTable(table); err != nil {
		return "", err
	}

	if s.db != nil {
		// Use database-specific placeholder: $1 for postgres, ? for mysql/sqlite3
		var ph string
		switch s.Driver {
		case "postgres":
			ph = "$1"
		default:
			ph = "?"
		}
		q := fmt.Sprintf("SELECT * FROM %s WHERE id = %s LIMIT 1", quoteIdent(table), ph)
		rows, err := s.db.Query(q, id)
		if err != nil {
			return "", fmt.Errorf("find query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	}

	// CLI fallback — ID is not directly injectable since table is validated
	q := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s' LIMIT 1;", quoteIdent(table), escapeSingle(id))
	return s.ExecFormatted(q)
}
