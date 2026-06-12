package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// Describe shows the schema of a table. Table name is validated against the database.
func (s *Session) Describe(table string) (string, error) {
	if err := s.validateTable(table); err != nil {
		return "", err
	}
	if s.db != nil {
		return s.describeNative(table)
	}
	return s.describeCLI(table)
}

func (s *Session) validateTable(table string) error {
	// Reject obviously malicious input
	if strings.Contains(table, ";") || strings.Contains(table, "--") || strings.Contains(table, "/*") {
		return fmt.Errorf("invalid table name: %q", table)
	}
	// Check for spaces (unlikely in real table names, common in injection)
	if strings.ContainsAny(table, " \t\n\r") {
		return fmt.Errorf("invalid table name: %q", table)
	}

	// If we have a native connection, verify the table actually exists
	if s.db != nil {
		var q string
		var row *sql.Row
		switch s.Type {
		case "postgres":
			q = "SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = $1"
			row = s.db.QueryRow(q, table)
		case "mysql":
			q = "SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
			row = s.db.QueryRow(q, table)
		case "sqlite3":
			q = "SELECT 1 FROM sqlite_master WHERE type='table' AND name = ?"
			row = s.db.QueryRow(q, table)
		default:
			return nil // can't validate, allow through
		}
		var exists int
		if err := row.Scan(&exists); err != nil {
			return fmt.Errorf("table %q not found", table)
		}
	}
	return nil
}

func (s *Session) describeNative(table string) (string, error) {
	switch s.Type {
	case "postgres":
		rows, err := s.db.Query(
			"SELECT column_name as \"Column\", data_type as \"Type\", is_nullable as \"Nullable\", column_default as \"Default\" FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position",
			table,
		)
		if err != nil {
			return "", fmt.Errorf("describe query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	case "sqlite3":
		rows, err := s.db.Query(
			fmt.Sprintf("SELECT name as \"Column\", type as \"Type\", CASE WHEN \"notnull\" = 1 THEN 'NOT NULL' ELSE 'NULL' END as \"Nullable\", \"dflt_value\" as \"Default\", CASE WHEN pk > 0 THEN 'PK' ELSE '' END as \"Key\" FROM pragma_table_info(%s) ORDER BY cid", quoteStr(table)),
		)
		if err != nil {
			return "", fmt.Errorf("describe query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	case "mysql":
		rows, err := s.db.Query(
			"SELECT column_name as \"Column\", column_type as \"Type\", is_nullable as \"Nullable\", column_key as \"Key\", column_default as \"Default\", extra as \"Extra\" FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? ORDER BY ordinal_position",
			table,
		)
		if err != nil {
			return "", fmt.Errorf("describe query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	default:
		return s.describeCLI(table)
	}
}

func (s *Session) describeCLI(table string) (string, error) {
	var q string
	switch s.Type {
	case "postgres":
		q = fmt.Sprintf("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position;", escapeSingle(table))
	case "mysql":
		q = fmt.Sprintf("DESCRIBE %s;", quoteIdent(table))
	case "sqlite3":
		q = fmt.Sprintf("SELECT name as \"Column\", type as \"Type\", CASE WHEN \"notnull\" = 1 THEN 'NOT NULL' ELSE 'NULL' END as \"Nullable\", \"dflt_value\" as \"Default\", CASE WHEN pk > 0 THEN 'PK' ELSE '' END as \"Key\" FROM pragma_table_info(%s) ORDER BY cid;", quoteStr(table))
	default:
		q = fmt.Sprintf("SELECT * FROM %s LIMIT 0;", quoteIdent(table))
	}
	return s.ExecFormatted(q)
}
