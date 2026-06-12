package db

import (
	"fmt"
	"strings"
)

// ExploreTable returns table data for the explore TUI.
// It fetches column info and a limited number of rows.
func (s *Session) ExploreTable(table string, limit int) ([]string, [][]string, error) {
	if err := s.validateTable(table); err != nil {
		return nil, nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	q := fmt.Sprintf("SELECT * FROM %s LIMIT %d", quoteIdent(table), limit) //nolint:gosec // table is validated and quoteIdent-sanitized

	if s.db != nil {
		rows, err := s.db.Query(q)
		if err != nil {
			return nil, nil, fmt.Errorf("query: %w", err)
		}
		defer rows.Close()
		return readRows(rows)
	}

	// CLI fallback
	out, err := s.ExecFormatted(q)
	if err != nil {
		return nil, nil, err
	}
	headers, rows := parseCLIOutput(out)
	if len(headers) == 0 {
		return nil, nil, fmt.Errorf("no columns returned — table may be empty or output could not be parsed")
	}
	return headers, rows, nil
}

// ExploreRowCount returns the total number of rows in a table.
func (s *Session) ExploreRowCount(table string) (int64, error) {
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(table)) //nolint:gosec // table is validated and quoteIdent-sanitized

	if s.db != nil {
		var count int64
		err := s.db.QueryRow(q).Scan(&count)
		return count, err
	}

	// CLI fallback
	out, err := s.Exec(q + ";")
	if err != nil {
		return 0, err
	}
	var count int64
	if _, err := fmt.Sscanf(strings.TrimSpace(out), "%d", &count); err != nil {
		return 0, fmt.Errorf("parsing row count: %w", err)
	}
	return count, nil
}

// readRows reads column headers and row data from sql.Rows.
func readRows(rows interface {
	Columns() ([]string, error)
	Next() bool
	Scan(...interface{}) error
}) ([]string, [][]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var data [][]string
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make([]string, len(cols))
		for i, v := range values {
			if v == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		data = append(data, row)
	}

	return cols, data, nil
}

// parseCLIOutput parses CLI tabular output into headers and rows.
func parseCLIOutput(out string) ([]string, [][]string) {
	lines := strings.Split(out, "\n")
	var nonEmpty []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "-") {
			nonEmpty = append(nonEmpty, l)
		}
	}
	if len(nonEmpty) == 0 {
		return nil, nil
	}

	// First line is header
	headers := splitRow(nonEmpty[0])
	var rows [][]string
	for _, l := range nonEmpty[1:] {
		rows = append(rows, splitRow(l))
	}
	return headers, rows
}

func splitRow(line string) []string {
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
