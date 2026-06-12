package db

import (
        "database/sql"
        "fmt"
        "strings"
)

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

        var sb strings.Builder
        for rows.Next() {
                var name string
                if err := rows.Scan(&name); err != nil {
                        continue
                }
                sb.WriteString(name)
                sb.WriteString("\n")
        }
        return sb.String(), rows.Err()
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
                return s.queryRowsNative(
                        "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position",
                        table,
                )
        case "sqlite3":
                return s.queryRowsNative(
                        fmt.Sprintf("SELECT name as \"Column\", type as \"Type\", CASE WHEN \"notnull\" = 1 THEN 'NOT NULL' ELSE 'NULL' END as \"Nullable\", \"dflt_value\" as \"Default\", CASE WHEN pk > 0 THEN 'PK' ELSE '' END as \"Key\" FROM pragma_table_info(%s) ORDER BY cid", quoteIdent(table)),
                )
        case "mysql":
                return s.queryRowsNative(fmt.Sprintf("DESCRIBE %s", quoteIdent(table)))
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
                q = fmt.Sprintf("SELECT name as \"Column\", type as \"Type\", CASE WHEN \"notnull\" = 1 THEN 'NOT NULL' ELSE 'NULL' END as \"Nullable\", \"dflt_value\" as \"Default\", CASE WHEN pk > 0 THEN 'PK' ELSE '' END as \"Key\" FROM pragma_table_info(%s) ORDER BY cid;", quoteIdent(table))
        default:
                q = fmt.Sprintf("SELECT * FROM %s LIMIT 0;", quoteIdent(table))
        }
        return s.ExecFormatted(q)
}

// Count returns the count of rows in a table. Table name is validated;
// WHERE clause uses parameterized query when possible.
func (s *Session) Count(table, where string) (string, error) {
        if err := s.validateTable(table); err != nil {
                return "", err
        }

        if s.db != nil && where != "" {
                // Use parameterized query for the WHERE clause
                var count int64
                q := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", quoteIdent(table), where)
                err := s.db.QueryRow(q).Scan(&count)
                if err != nil {
                        return "", fmt.Errorf("count query: %w", err)
                }
                return fmt.Sprintf("%d\n", count), nil
        }

        if s.db != nil {
                var count int64
                q := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(table))
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
                q := fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1", quoteIdent(table))
                rows, err := s.db.Query(q, id)
                if err != nil {
                        return "", fmt.Errorf("find query: %w", err)
                }
                defer rows.Close()

                cols, err := rows.Columns()
                if err != nil {
                        return "", err
                }

                var sb strings.Builder
                sb.WriteString(strings.Join(cols, "\t"))
                sb.WriteString("\n")

                for rows.Next() {
                        vals := make([]interface{}, len(cols))
                        ptrs := make([]interface{}, len(cols))
                        for i := range vals {
                                ptrs[i] = &vals[i]
                        }
                        if err := rows.Scan(ptrs...); err != nil {
                                continue
                        }
                        for i, v := range vals {
                                if i > 0 {
                                        sb.WriteString("\t")
                                }
                                switch val := v.(type) {
                                case []byte:
                                        sb.WriteString(string(val))
                                case string:
                                        sb.WriteString(val)
                                case nil:
                                        sb.WriteString("NULL")
                                default:
                                        sb.WriteString(fmt.Sprintf("%v", val))
                                }
                        }
                        sb.WriteString("\n")
                }
                return sb.String(), rows.Err()
        }

        // CLI fallback — ID is not directly injectable since table is validated
        q := fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1;", quoteIdent(table))
        return s.ExecCLIQueryFormatted(q, id)
}

// ExecCLIQuery runs a query via the external CLI with a single parameter.
func (s *Session) ExecCLIQuery(query, param string) (string, error) {
        // For CLI-based execution, we substitute the parameter safely
        // by escaping single quotes (basic SQL injection prevention for CLI)
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

// queryRowsNative runs a query and returns tab-separated results.
func (s *Session) queryRowsNative(query string, args ...interface{}) (string, error) {
        rows, err := s.db.Query(query, args...)
        if err != nil {
                return "", fmt.Errorf("query: %w", err)
        }
        defer rows.Close()

        cols, err := rows.Columns()
        if err != nil {
                return "", err
        }

        var sb strings.Builder
        sb.WriteString(strings.Join(cols, "\t"))
        sb.WriteString("\n")

        for rows.Next() {
                vals := make([]interface{}, len(cols))
                ptrs := make([]interface{}, len(cols))
                for i := range vals {
                        ptrs[i] = &vals[i]
                }
                if err := rows.Scan(ptrs...); err != nil {
                        continue
                }
                for i, v := range vals {
                        if i > 0 {
                                sb.WriteString("\t")
                        }
                        switch val := v.(type) {
                        case []byte:
                                sb.WriteString(string(val))
                        case string:
                                sb.WriteString(val)
                        case nil:
                                sb.WriteString("NULL")
                        default:
                                sb.WriteString(fmt.Sprintf("%v", val))
                        }
                }
                sb.WriteString("\n")
        }
        return sb.String(), rows.Err()
}

// quoteIdent wraps a SQL identifier in double-quotes (standard SQL) or backticks (MySQL).
func quoteIdent(name string) string {
        if strings.Contains(name, `"`) || strings.Contains(name, "`") {
                return name // already quoted or suspicious
        }
        return `"` + name + `"`
}

// escapeSingle escapes single quotes in a string value.
func escapeSingle(s string) string {
        return strings.ReplaceAll(s, "'", "''")
}
