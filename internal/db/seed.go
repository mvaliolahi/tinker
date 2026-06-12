package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Seed executes SQL seed files against the database.
// It reads all .sql files from the given directory (or a single file),
// splits them by semicolons, and executes each statement.
func (s *Session) Seed(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("seed path %q not found: %w", path, err)
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("reading seed directory: %w", err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
				files = append(files, filepath.Join(path, e.Name()))
			}
		}
		if len(files) == 0 {
			return fmt.Errorf("no .sql seed files found in %s", path)
		}
	} else {
		files = []string{path}
	}

	return s.execFiles(files, "seed")
}

// execFiles reads and executes each SQL file in order.
func (s *Session) execFiles(files []string, label string) error {
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("reading %s file %s: %w", label, f, err)
		}

		stmts := splitSQL(string(data))
		if len(stmts) == 0 {
			continue
		}

		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if err := s.execSQL(stmt); err != nil {
				return fmt.Errorf("%s file %s: %w", label, filepath.Base(f), err)
			}
		}
	}
	return nil
}

// execSQL runs a single SQL statement via native connection or CLI.
func (s *Session) execSQL(stmt string) error {
	if s.db != nil {
		_, err := s.db.Exec(stmt)
		return err
	}
	// CLI fallback
	_, err := s.Exec(stmt)
	return err
}

// splitSQL splits SQL content by semicolons, respecting single-quoted strings.
func splitSQL(content string) []string {
	var stmts []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(content); i++ {
		ch := content[i]

		if ch == '\'' && !inQuote {
			inQuote = true
			current.WriteByte(ch)
			continue
		}
		if ch == '\'' && inQuote {
			// Check for escaped quote ''
			if i+1 < len(content) && content[i+1] == '\'' {
				current.WriteString("''")
				i++
				continue
			}
			inQuote = false
			current.WriteByte(ch)
			continue
		}

		if ch == ';' && !inQuote {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			current.Reset()
			continue
		}

		// Skip line comments (-- ...)
		if ch == '-' && !inQuote && i+1 < len(content) && content[i+1] == '-' {
			// Skip to end of line
			for i < len(content) && content[i] != '\n' {
				i++
			}
			continue
		}

		current.WriteByte(ch)
	}

	// Add any remaining statement (without trailing semicolon)
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		stmts = append(stmts, stmt)
	}

	return stmts
}
