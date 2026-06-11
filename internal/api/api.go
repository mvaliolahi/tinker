// Package api provides HTTP API interaction capabilities by wrapping curl/httpie.
// It resolves connection details from tinker.toml and enables calling API
// endpoints from the terminal with automatic auth and base URL resolution.
package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/mvaliolahi/tinker/internal/config"
)

// Session represents an API session configuration.
type Session struct {
	BaseURL string
	Auth    string
	AuthType string
	Headers map[string]string
	Spec    string
}

// NewSession creates a new API session from config.
func NewSession(cfg *config.APIConfig) (*Session, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no api configuration found in tinker.toml")
	}

	if cfg.ResolvedBaseURL == "" {
		return nil, fmt.Errorf("api base_url is empty after resolving")
	}

	return &Session{
		BaseURL:  strings.TrimRight(cfg.ResolvedBaseURL, "/"),
		Auth:     cfg.ResolvedAuth,
		AuthType: cfg.AuthType,
		Headers:  cfg.Headers,
		Spec:     cfg.Spec,
	}, nil
}

// Request makes an HTTP request with the configured auth and headers.
func (s *Session) Request(method, path string, body string, extraHeaders map[string]string) (string, error) {
	url := s.buildURL(path)

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Set default headers
	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}

	// Set auth header
	if s.Auth != "" {
		switch strings.ToLower(s.AuthType) {
		case "bearer":
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Auth))
		case "basic":
			req.Header.Set("Authorization", fmt.Sprintf("Basic %s", s.Auth))
		case "api_key":
			req.Header.Set("X-API-Key", s.Auth)
		default:
			req.Header.Set("Authorization", s.Auth)
		}
	}

	// Set extra headers
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	// Set Content-Type if body is present and not set
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status))
	for k, v := range resp.Header {
		buf.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
	}
	buf.WriteString("\n")
	buf.Write(respBody)

	return buf.String(), nil
}

// Get makes a GET request.
func (s *Session) Get(path string) (string, error) {
	return s.Request("GET", path, "", nil)
}

// Post makes a POST request.
func (s *Session) Post(path, body string) (string, error) {
	return s.Request("POST", path, body, nil)
}

// Put makes a PUT request.
func (s *Session) Put(path, body string) (string, error) {
	return s.Request("PUT", path, body, nil)
}

// Delete makes a DELETE request.
func (s *Session) Delete(path string) (string, error) {
	return s.Request("DELETE", path, "", nil)
}

// Interactive opens an interactive HTTPie-like session if httpie is installed,
// otherwise falls back to raw request mode.
func (s *Session) Interactive() error {
	// Try httpie first
	if _, err := exec.LookPath("httpl"); err == nil {
		return s.httpieSession()
	}
	if _, err := exec.LookPath("http"); err == nil {
		return s.httpieSession()
	}

	// Fall back to curlie
	if _, err := exec.LookPath("curlie"); err == nil {
		return s.curlieSession()
	}

	return fmt.Errorf("no interactive HTTP client found. Install one of:\n  httpie: pip install httpie or brew install httpie\n  curlie: go install github.com/rs/curlie@latest")
}

// httpieSession opens an httpie session with configured base URL and auth.
func (s *Session) httpieSession() error {
	args := []string{"--session", "tinker"}

	if s.Auth != "" {
		switch strings.ToLower(s.AuthType) {
		case "bearer":
			args = append(args, fmt.Sprintf("Authorization:Bearer %s", s.Auth))
		case "api_key":
			args = append(args, fmt.Sprintf("X-API-Key:%s", s.Auth))
		}
	}

	args = append(args, s.BaseURL)

	cmd := exec.Command("http", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// curlieSession opens a curlie session.
func (s *Session) curlieSession() error {
	args := []string{s.BaseURL}

	if s.Auth != "" {
		switch strings.ToLower(s.AuthType) {
		case "bearer":
			args = append([]string{"-H", fmt.Sprintf("Authorization: Bearer %s", s.Auth)}, args...)
		case "api_key":
			args = append([]string{"-H", fmt.Sprintf("X-API-Key: %s", s.Auth)}, args...)
		}
	}

	cmd := exec.Command("curlie", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildURL constructs the full URL from base URL and path.
func (s *Session) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	path = strings.TrimLeft(path, "/")
	return fmt.Sprintf("%s/%s", s.BaseURL, path)
}
