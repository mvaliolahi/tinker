package api

import (
	"fmt"
	"os"
	"os/exec"
)

func (s *Session) Interactive() error {
	if p, _ := exec.LookPath("http"); p != "" {
		return s.runHTTPie()
	}
	if p, _ := exec.LookPath("curlie"); p != "" {
		return s.runCurlie()
	}
	return fmt.Errorf("no HTTP client found — install httpie or curlie")
}

func (s *Session) runHTTPie() error {
	args := []string{"--session", "tinker"}

	for k, v := range s.authHeaders() {
		args = append(args, fmt.Sprintf("%s:%s", k, v))
	}
	args = append(args, s.BaseURL)

	cmd := exec.Command("http", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func (s *Session) runCurlie() error {
	args := []string{s.BaseURL}

	for k, v := range s.authHeaders() {
		args = append([]string{"-H", fmt.Sprintf("%s: %s", k, v)}, args...)
	}

	cmd := exec.Command("curlie", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
