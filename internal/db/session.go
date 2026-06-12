package db

import (
        "database/sql"
        "fmt"
        "strings"

        "github.com/mvaliolahi/tinker/internal/config"
)

// Session holds database connection info and an optional native database/sql connection.
type Session struct {
        Driver string
        DSN    string
        Type   string
        db     *sql.DB // native connection for safe queries (nil if unavailable)
}

// NewSession creates a new database session from config.
func NewSession(cfg *config.Database) (*Session, error) {
        if cfg == nil {
                return nil, fmt.Errorf("no [database] section in tinker.toml")
        }
        if cfg.URL == "" {
                return nil, fmt.Errorf("database URL is empty (source: %s)", cfg.Source)
        }

        driver := cfg.Driver
        if driver == "" {
                driver = driverFromType(cfg.Type)
        }

        s := &Session{Driver: driver, DSN: cfg.URL, Type: cfg.Type}

        // Try to open a native database/sql connection for safe parameterized queries.
        // This is best-effort; interactive sessions still use external CLIs.
        s.tryOpenNative()

        return s, nil
}

// tryOpenNative opens a database/sql connection if the driver is available.
func (s *Session) tryOpenNative() {
        dsn := s.DSN
        driver := s.Driver

        // Adjust DSN for sqlite3: needs file: prefix or raw path
        if driver == "sqlite3" {
                if !strings.HasPrefix(dsn, "file:") && !strings.HasPrefix(dsn, "/") {
                        dsn = "file:" + dsn + "?mode=ro"
                } else if strings.HasPrefix(dsn, "/") {
                        dsn = "file:" + dsn + "?mode=ro"
                } else if strings.HasPrefix(dsn, "file:") && !strings.Contains(dsn, "?") {
                        // file:path without existing query params — add mode=ro
                        dsn = dsn + "?mode=ro"
                }
                // If file:path?params already exists, don't add mode=ro (user may want rw)
        }

        db, err := sql.Open(driver, dsn)
        if err != nil {
                return
        }
        // Verify the connection actually works
        if err := db.Ping(); err != nil {
                db.Close()
                return
        }
        s.db = db
}

// Close closes the native database connection if open.
func (s *Session) Close() {
        if s.db != nil {
                s.db.Close()
        }
}

// HasNativeConn returns true if a native database/sql connection is available.
func (s *Session) HasNativeConn() bool {
        return s.db != nil
}

func driverFromType(t string) string {
        switch strings.ToLower(t) {
        case "postgres", "postgresql":
                return "postgres"
        case "mysql":
                return "mysql"
        case "sqlite3", "sqlite":
                return "sqlite3"
        case "sqlserver", "mssql":
                return "sqlserver"
        case "mongodb", "mongo":
                return "mongo"
        default:
                return t
        }
}
