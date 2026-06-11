package api

import (
	"fmt"
	"strings"

	"github.com/mvaliolahi/tinker/internal/config"
)

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

	base := strings.TrimRight(cfg.ResolvedBaseURL, "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}

	return &Session{
		BaseURL:  base,
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
