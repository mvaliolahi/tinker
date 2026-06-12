package db

import "fmt"

// Tables lists all tables in the database using safe parameterized queries
// when a native connection is available, falling back to CLI otherwise.
func (s *Session) Tables() (string, error) {
	if s.db != nil {
		return s.tablesNative()
	}
	return s.tablesCLI()
}

func (s *Session) tablesNative() (string, error) {
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
		return "", fmt.Errorf("querying tables: %w", err)
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
	return renderSimpleList(names), rows.Err()
}

func (s *Session) tablesCLI() (string, error) {
	var q string
	switch s.Type {
	case "postgres":
		q = "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;"
	case "mysql":
		q = "SHOW TABLES;"
	case "sqlite3":
		q = "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;"
	default:
		q = "SELECT table_name FROM information_schema.tables ORDER BY table_name;"
	}
	return s.Exec(q)
}
