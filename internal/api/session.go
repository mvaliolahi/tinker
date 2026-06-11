package api

import (
	"fmt"
	"strings"

	"github.com/mvaliolahi/tinker/internal/config"
)

func normalizeBaseURL(raw string) string {
	base := strings.TrimRight(raw, "/")

	if strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://") {
		return base
	}

	if strings.HasPrefix(base, ":") {
		base = "localhost" + base
	} else if isPortOnly(base) {
		base = "localhost:" + base
	}

	return "http://" + base
}

func isPortOnly(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

type Session struct {
	BaseURL  string
	Auth     string
	AuthType string
	Headers  map[string]string
	Spec     string
}

func NewSession(cfg *config.API) (*Session, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no [api] section in tinker.toml")
	}
	if cfg.ResolvedBaseURL == "" {
		return nil, fmt.Errorf("api base_url is empty")
	}

	return &Session{
		BaseURL:  normalizeBaseURL(cfg.ResolvedBaseURL),
		Auth:     cfg.ResolvedAuth,
		AuthType: cfg.AuthType,
		Headers:  cfg.Headers,
		Spec:     cfg.Spec,
	}, nil
}

func (s *Session) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return s.BaseURL + "/" + strings.TrimLeft(path, "/")
}

func (s *Session) authHeaders() map[string]string {
	h := make(map[string]string)
	if s.Auth == "" {
		return h
	}
	switch strings.ToLower(s.AuthType) {
	case "bearer":
		h["Authorization"] = "Bearer " + s.Auth
	case "basic":
		h["Authorization"] = "Basic " + s.Auth
	case "api_key":
		h["X-API-Key"] = s.Auth
	default:
		h["Authorization"] = s.Auth
	}
	return h
}
