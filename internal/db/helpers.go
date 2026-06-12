package db

import "strings"

// quoteIdent wraps a SQL identifier in double-quotes (standard SQL) or backticks (MySQL).
func quoteIdent(name string) string {
	// Strip any existing quotes first to prevent bypass
	clean := strings.Trim(strings.Trim(name, `"`), "`")
	return `"` + clean + `"`
}

// quoteStr wraps a string value in single-quotes for SQLite pragma functions.
// pragma_table_info('users') expects a string literal, not a double-quoted identifier.
func quoteStr(name string) string {
	clean := strings.Trim(strings.Trim(name, `"`), "'")
	return "'" + escapeSingle(clean) + "'"
}

// escapeSingle escapes single quotes in a string value.
func escapeSingle(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
