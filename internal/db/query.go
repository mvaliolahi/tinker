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

// Ping tests database connectivity. Returns nil on success.
func (s *Session) Ping() error {
        if s.db != nil {
                return s.db.Ping()
        }
        // No native connection — try a trivial CLI query
        _, err := s.Exec("SELECT 1;")
        return err
}

// Size returns row counts for all tables.
func (s *Session) Size() (string, error) {
        if s.db != nil {
                return s.sizeNative()
        }
        return s.sizeCLI()
}

func (s *Session) sizeNative() (string, error) {
        switch s.Type {
        case "sqlite3":
                // Get all table names and count each one
                names, err := s.tableNamesNative()
                if err != nil {
                        return "", err
                }
                return s.countTablesRendered(names)
        case "postgres":
                rows, err := s.db.Query(
                        "SELECT relname as \"Table\", n_live_tup as \"Rows\" FROM pg_stat_user_tables ORDER BY n_live_tup DESC;",
                )
                if err != nil {
                        return "", fmt.Errorf("size query: %w", err)
                }
                defer rows.Close()
                return renderTableFromRows(rows)
        case "mysql":
                rows, err := s.db.Query(
                        "SELECT table_name as \"Table\", table_rows as \"Rows\" FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_rows DESC;",
                )
                if err != nil {
                        return "", fmt.Errorf("size query: %w", err)
                }
                defer rows.Close()
                return renderTableFromRows(rows)
        default:
                return s.sizeCLI()
        }
}

// tableNamesNative returns just the table names as a slice (no formatting).
func (s *Session) tableNamesNative() ([]string, error) {
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
                return nil, fmt.Errorf("querying tables: %w", err)
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
        return names, rows.Err()
}

// countTablesRendered counts rows in each table and renders a go-pretty table.
func (s *Session) countTablesRendered(tableNames []string) (string, error) {
        headers := []string{"Table", "Rows"}
        var dataRows [][]string

        for _, t := range tableNames {
                var count int64
                q := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(t))
                if err := s.db.QueryRow(q).Scan(&count); err != nil {
                        dataRows = append(dataRows, []string{t, "?"})
                        continue
                }
                dataRows = append(dataRows, []string{t, fmt.Sprintf("%d", count)})
        }

        return renderTable(headers, dataRows), nil
}

func (s *Session) sizeCLI() (string, error) {
        var q string
        switch s.Type {
        case "sqlite3":
                // CLI fallback: just list tables — counting via CLI is impractical
                return s.Tables()
        case "postgres":
                q = "SELECT relname as \"Table\", n_live_tup as \"Rows\" FROM pg_stat_user_tables ORDER BY n_live_tup DESC;"
        case "mysql":
                q = "SELECT table_name as \"Table\", table_rows as \"Rows\" FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_rows DESC;"
        default:
                q = "SELECT table_name, table_rows FROM information_schema.tables ORDER BY table_rows DESC;"
        }
        return s.ExecFormatted(q)
}

// Count returns the count of rows in a table. Table name is validated;
// WHERE clause is treated as a raw SQL condition (user is responsible for safety).
func (s *Session) Count(table, where string) (string, error) {
        if err := s.validateTable(table); err != nil {
                return "", err
        }

        if s.db != nil {
                var count int64
                var q string
                if where != "" {
                        q = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", quoteIdent(table), where)
                } else {
                        q = fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(table))
                }
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
                // Use database-specific placeholder: $1 for postgres, ? for mysql/sqlite3
                var ph string
                switch s.Driver {
                case "postgres":
                        ph = "$1"
                default:
                        ph = "?"
                }
                q := fmt.Sprintf("SELECT * FROM %s WHERE id = %s LIMIT 1", quoteIdent(table), ph)
                rows, err := s.db.Query(q, id)
                if err != nil {
                        return "", fmt.Errorf("find query: %w", err)
                }
                defer rows.Close()
                return renderTableFromRows(rows)
        }

        // CLI fallback — ID is not directly injectable since table is validated
        q := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s' LIMIT 1;", quoteIdent(table), escapeSingle(id))
        return s.ExecFormatted(q)
}

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
