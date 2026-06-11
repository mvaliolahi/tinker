package api

import (
	"fmt"
	"os/exec"

	"github.com/mvaliolahi/tinker/internal/runner"
)

func (s *Session) Interactive() error {
	if p, _ := exec.LookPath("http"); p != "" {
		return s.runHTTPie()
	}
	if p, _ := exec.LookPath("curlie"); p != "" {
		return s.runCurlie()
	}
	return fmt.Errorf("no HTTP client found — run 'tinker deps' to install curlie")
}

func (s *Session) runHTTPie() error {
	args := []string{"--session", "tinker"}

	for k, v := range s.authHeaders() {
		args = append(args, fmt.Sprintf("%s:%s", k, v))
	}
	args = append(args, s.BaseURL)

	return runner.Interactive("http", args...)
}

func (s *Session) runCurlie() error {
	args := []string{s.BaseURL}

	for k, v := range s.authHeaders() {
		args = append([]string{"-H", fmt.Sprintf("%s: %s", k, v)}, args...)
	}

	return runner.Interactive("curlie", args...)
}
