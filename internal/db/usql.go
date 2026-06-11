package db

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mvaliolahi/tinker/internal/runner"
)

func (s *Session) dsnForUSQL() string {
	dsn := s.DSN
	if !strings.Contains(dsn, "://") && !strings.HasPrefix(dsn, "/") {
		dsn = fmt.Sprintf("%s://%s", s.Driver, dsn)
	}
	return dsn
}

func (s *Session) Connect() error {
	if _, err := exec.LookPath("usql"); err != nil {
		return fmt.Errorf("usql not found — run 'tinker deps' to install")
	}

	return runner.Interactive("usql", s.dsnForUSQL())
}

func (s *Session) Exec(query string) (string, error) {
	if _, err := exec.LookPath("usql"); err != nil {
		return "", fmt.Errorf("usql not found — run 'tinker deps' to install")
	}

	cmd := exec.Command("usql", s.dsnForUSQL(), "-c", query)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
