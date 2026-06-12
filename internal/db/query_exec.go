package db

import (
	"fmt"
	"strings"
)

// ExecNative runs a raw SQL query via the native connection and returns
// formatted table results using go-pretty. Returns ("", nil) if no native connection.
func (s *Session) ExecNative(query string) (string, error) {
	if s.db == nil {
		return "", nil
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	return renderTableFromRows(rows)
}

// ExecCLIQuery runs a query via the external CLI with a single parameter.
func (s *Session) ExecCLIQuery(query, param string) (string, error) {
	safe := escapeSingle(param)
	query = strings.Replace(query, "?", "'"+safe+"'", 1)
	return s.Exec(query)
}

// ExecCLIQueryFormatted runs a formatted query via the external CLI with a single parameter.
func (s *Session) ExecCLIQueryFormatted(query, param string) (string, error) {
	safe := escapeSingle(param)
	query = strings.Replace(query, "?", "'"+safe+"'", 1)
	return s.ExecFormatted(query)
}
