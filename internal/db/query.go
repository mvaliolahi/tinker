package db

import "fmt"

func (s *Session) Tables() (string, error) {
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

func (s *Session) Describe(table string) (string, error) {
	var q string
	switch s.Type {
	case "postgres":
		q = fmt.Sprintf("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position;", table)
	case "mysql":
		q = fmt.Sprintf("DESCRIBE %s;", table)
	case "sqlite3":
		q = fmt.Sprintf("PRAGMA table_info(%s);", table)
	default:
		q = fmt.Sprintf("SELECT * FROM %s LIMIT 0;", table)
	}
	return s.Exec(q)
}

func (s *Session) Count(table, where string) (string, error) {
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if where != "" {
		q += fmt.Sprintf(" WHERE %s", where)
	}
	return s.Exec(q + ";")
}

func (s *Session) Find(table, id string) (string, error) {
	return s.Exec(fmt.Sprintf("SELECT * FROM %s WHERE id = %s LIMIT 1;", table, id))
}
