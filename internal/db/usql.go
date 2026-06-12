package db

import (
        "fmt"
        "os/exec"
        "strings"

        "github.com/mvaliolahi/tinker/internal/runner"
)

func (s *Session) CLI() string {
        // 1. Modern CLI with syntax highlighting + autocomplete
        if cli := s.modernCLI(); cli != "" {
                if _, err := exec.LookPath(cli); err == nil {
                        return cli
                }
        }
        // 2. Native CLI
        if cli, _ := s.nativeConnect(); cli != "" {
                if _, err := exec.LookPath(cli); err == nil {
                        return cli
                }
        }
        return "usql"
}

// modernCLI returns the name of a modern DB REPL with syntax highlighting + autocomplete.
func (s *Session) modernCLI() string {
        switch s.Driver {
        case "sqlite3":
                return "litecli"
        case "postgres":
                return "pgcli"
        case "mysql":
                return "mycli"
        default:
                return ""
        }
}

// ModernCLIAvailable returns the modern CLI name for the current driver
// (even if not installed), used for hint messages.
func (s *Session) ModernCLIAvailable() string {
        return s.modernCLI()
}

func (s *Session) dsnForUSQL() string {
        dsn := s.DSN
        if !strings.Contains(dsn, "://") && !strings.HasPrefix(dsn, "/") {
                dsn = fmt.Sprintf("%s://%s", s.Driver, dsn)
        }
        return dsn
}

func (s *Session) Connect() error {
        // 1. Try modern CLI first (syntax highlighting + autocomplete)
        if cli, args := s.modernConnect(); cli != "" {
                if _, err := exec.LookPath(cli); err == nil {
                        return runner.Interactive(cli, args...)
                }
        }

        // 2. Try native CLI
        if cli, args := s.nativeConnect(); cli != "" {
                if _, err := exec.LookPath(cli); err == nil {
                        return runner.Interactive(cli, args...)
                }
        }

        // 3. usql fallback
        if _, err := exec.LookPath("usql"); err != nil {
                return fmt.Errorf("no database CLI found — run 'tinker deps install'")
        }

        return runner.Interactive("usql", s.dsnForUSQL())
}

// modernConnect returns the modern CLI name and args for an interactive session.
func (s *Session) modernConnect() (string, []string) {
        switch s.Driver {
        case "sqlite3":
                return "litecli", []string{s.dsnForCLI()}
        case "postgres":
                return "pgcli", []string{s.DSN}
        case "mysql":
                return "mycli", []string{s.DSN}
        default:
                return "", nil
        }
}

// dsnForCLI returns the DSN in the format expected by the native CLI tool.
// For sqlite3, this strips the "file:" prefix since the sqlite3 CLI expects a plain path.
func (s *Session) dsnForCLI() string {
        if s.Driver == "sqlite3" {
                dsn := s.DSN
                // Strip file: prefix and query params for sqlite3 CLI
                dsn = strings.TrimPrefix(dsn, "file:")
                if idx := strings.Index(dsn, "?"); idx != -1 {
                        dsn = dsn[:idx]
                }
                return dsn
        }
        return s.DSN
}

func (s *Session) nativeConnect() (string, []string) {
        switch s.Driver {
        case "sqlite3":
                return "sqlite3", []string{s.dsnForCLI()}
        case "postgres":
                return "psql", []string{s.DSN}
        case "mysql":
                return "mysql", []string{s.DSN}
        default:
                return "", nil
        }
}

func (s *Session) Exec(query string) (string, error) {
        cli, args := s.nativeExec(query, false)
        if cli != "" {
                if _, err := exec.LookPath(cli); err == nil {
                        out, err := exec.Command(cli, args...).CombinedOutput()
                        return string(out), err
                }
        }

        if _, err := exec.LookPath("usql"); err != nil {
                return "", fmt.Errorf("usql not found — run 'tinker deps' to install")
        }

        cmd := exec.Command("usql", s.dsnForUSQL(), "-c", query)
        out, err := cmd.CombinedOutput()
        return string(out), err
}

// ExecFormatted runs a query with tabular formatting (headers + column alignment).
// Use this for describe, find, and any result set where readability matters.
func (s *Session) ExecFormatted(query string) (string, error) {
        cli, args := s.nativeExec(query, true)
        if cli != "" {
                if _, err := exec.LookPath(cli); err == nil {
                        out, err := exec.Command(cli, args...).CombinedOutput()
                        return string(out), err
                }
        }

        // usql always produces formatted output
        if _, err := exec.LookPath("usql"); err != nil {
                return "", fmt.Errorf("usql not found — run 'tinker deps' to install")
        }

        cmd := exec.Command("usql", s.dsnForUSQL(), "-c", query)
        out, err := cmd.CombinedOutput()
        return string(out), err
}

// nativeExec builds the CLI command for executing a query.
// When formatted is true, the command includes flags for tabular output (headers, column alignment).
func (s *Session) nativeExec(query string, formatted bool) (string, []string) {
        switch s.Driver {
        case "sqlite3":
                // sqlite3 [options] DB_PATH SQL
                if formatted {
                        return "sqlite3", []string{"-header", "-column", s.dsnForCLI(), query}
                }
                return "sqlite3", []string{s.dsnForCLI(), query}
        case "postgres":
                // psql always produces formatted output with -c
                return "psql", []string{s.DSN, "-c", query}
        case "mysql":
                // mysql always produces formatted output with -e
                return "mysql", []string{s.DSN, "-e", query}
        default:
                return "", nil
        }
}
