package db

import "fmt"

// Schema returns the CREATE TABLE statement for a table (with syntax highlighting).
func (s *Session) Schema(table string) (string, error) {
	if err := s.validateTable(table); err != nil {
		return "", err
	}
	if s.db != nil {
		return s.schemaNative(table)
	}
	return s.schemaCLI(table)
}

func (s *Session) schemaNative(table string) (string, error) {
	switch s.Type {
	case "sqlite3":
		var ddl string
		q := "SELECT sql FROM sqlite_master WHERE type='table' AND name = ?"
		if err := s.db.QueryRow(q, table).Scan(&ddl); err != nil {
			return "", fmt.Errorf("schema query: %w", err)
		}
		return ddl + ";\n", nil
	case "postgres":
		var ddl string
		q := "SELECT 'CREATE TABLE ' || tablename || ' (' || string_agg(column_name || ' ' || data_type || CASE WHEN is_nullable = 'NO' THEN ' NOT NULL' ELSE '' END || CASE WHEN column_default IS NOT NULL THEN ' DEFAULT ' || column_default ELSE '' END, ', ' ORDER BY ordinal_position) || ');' FROM information_schema.columns WHERE table_name = $1 GROUP BY tablename"
		if err := s.db.QueryRow(q, table).Scan(&ddl); err != nil {
			return "", fmt.Errorf("schema query: %w", err)
		}
		return ddl + "\n", nil
	case "mysql":
		rows, err := s.db.Query("SHOW CREATE TABLE " + quoteIdent(table))
		if err != nil {
			return "", fmt.Errorf("schema query: %w", err)
		}
		defer rows.Close()
		// SHOW CREATE TABLE returns: Table, Create Table
		if rows.Next() {
			var tableName, createStmt string
			if err := rows.Scan(&tableName, &createStmt); err != nil {
				return "", err
			}
			return createStmt + ";\n", nil
		}
		return "", fmt.Errorf("no schema found for table %s", table)
	default:
		return s.schemaCLI(table)
	}
}

func (s *Session) schemaCLI(table string) (string, error) {
	var q string
	switch s.Type {
	case "sqlite3":
		q = fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s';", escapeSingle(table))
	case "postgres":
		q = fmt.Sprintf("SELECT 'CREATE TABLE ' || tablename || ' (' || string_agg(column_name || ' ' || data_type || CASE WHEN is_nullable = 'NO' THEN ' NOT NULL' ELSE '' END, ', ' ORDER BY ordinal_position) || ');' FROM information_schema.columns WHERE table_name = '%s' GROUP BY tablename;", escapeSingle(table))
	case "mysql":
		q = fmt.Sprintf("SHOW CREATE TABLE %s;", quoteIdent(table))
	default:
		q = fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s';", escapeSingle(table))
	}
	return s.ExecFormatted(q)
}
