package db

import "fmt"

// Indexes shows the indexes on a table.
func (s *Session) Indexes(table string) (string, error) {
	if err := s.validateTable(table); err != nil {
		return "", err
	}
	if s.db != nil {
		return s.indexesNative(table)
	}
	return s.indexesCLI(table)
}

func (s *Session) indexesNative(table string) (string, error) {
	switch s.Type {
	case "postgres":
		rows, err := s.db.Query(
			"SELECT indexname as \"Name\", indexdef as \"Definition\" FROM pg_indexes WHERE tablename = $1 ORDER BY indexname",
			table,
		)
		if err != nil {
			return "", fmt.Errorf("indexes query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	case "sqlite3":
		rows, err := s.db.Query(
			fmt.Sprintf("SELECT il.name as \"Index\", CASE WHEN il.\"unique\" = 1 THEN 'UNIQUE' ELSE '' END as \"Unique\", il.origin as \"Origin\", GROUP_CONCAT(ii.name, ', ') as \"Columns\" FROM pragma_index_list(%s) il LEFT JOIN pragma_index_info(il.name) ii GROUP BY il.name ORDER BY il.name", quoteStr(table)),
		)
		if err != nil {
			return "", fmt.Errorf("indexes query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	case "mysql":
		rows, err := s.db.Query(
			"SELECT index_name as \"Name\", column_name as \"Column\", CASE WHEN non_unique = 0 THEN 'UNIQUE' ELSE '' END as \"Unique\", index_type as \"Type\" FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? ORDER BY index_name, seq_in_index",
			table,
		)
		if err != nil {
			return "", fmt.Errorf("indexes query: %w", err)
		}
		defer rows.Close()
		return renderTableFromRows(rows)
	default:
		return s.indexesCLI(table)
	}
}

func (s *Session) indexesCLI(table string) (string, error) {
	var q string
	switch s.Type {
	case "postgres":
		q = fmt.Sprintf("SELECT indexname as \"Name\", indexdef as \"Definition\" FROM pg_indexes WHERE tablename = '%s' ORDER BY indexname;", escapeSingle(table))
	case "sqlite3":
		q = fmt.Sprintf("SELECT name as \"Name\", CASE WHEN \"unique\" = 1 THEN 'UNIQUE' ELSE '' END as \"Unique\", origin as \"Origin\" FROM pragma_index_list(%s) ORDER BY name;", quoteStr(table))
	case "mysql":
		q = fmt.Sprintf("SHOW INDEX FROM %s;", quoteIdent(table))
	default:
		q = fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE tablename = '%s';", escapeSingle(table))
	}
	return s.ExecFormatted(q)
}
