package db

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		return fmt.Errorf("usql not found — install: go install github.com/xo/usql@latest")
	}

	cmd := exec.Command("usql", s.dsnForUSQL())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func (s *Session) Exec(query string) (string, error) {
	if _, err := exec.LookPath("usql"); err != nil {
		return "", fmt.Errorf("usql not found")
	}

	cmd := exec.Command("usql", s.dsnForUSQL(), "-c", query)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
