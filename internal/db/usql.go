package db

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mvaliolahi/tinker/internal/runner"
)

func (s *Session) CLI() string {
	if cli, _ := s.nativeConnect(); cli != "" {
		if _, err := exec.LookPath(cli); err == nil {
			return cli
		}
	}
	return "usql"
}

func (s *Session) dsnForUSQL() string {
	dsn := s.DSN
	if !strings.Contains(dsn, "://") && !strings.HasPrefix(dsn, "/") {
		dsn = fmt.Sprintf("%s://%s", s.Driver, dsn)
	}
	return dsn
}

func (s *Session) Connect() error {
	cli, args := s.nativeConnect()
	if cli != "" {
		if _, err := exec.LookPath(cli); err == nil {
			return runner.Interactive(cli, args...)
		}
	}

	if _, err := exec.LookPath("usql"); err != nil {
		return fmt.Errorf("usql not found — run 'tinker deps' to install")
	}

	return runner.Interactive("usql", s.dsnForUSQL())
}

func (s *Session) nativeConnect() (string, []string) {
	switch s.Driver {
	case "sqlite3":
		return "sqlite3", []string{s.DSN}
	case "postgres":
		return "psql", []string{s.DSN}
	case "mysql":
		return "mysql", []string{s.DSN}
	default:
		return "", nil
	}
}

func (s *Session) Exec(query string) (string, error) {
	cli, args := s.nativeExec(query)
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

func (s *Session) nativeExec(query string) (string, []string) {
	switch s.Driver {
	case "sqlite3":
		return "sqlite3", []string{s.DSN, "-c", query}
	case "postgres":
		return "psql", []string{s.DSN, "-c", query}
	case "mysql":
		return "mysql", []string{s.DSN, "-e", query}
	default:
		return "", nil
	}
}
