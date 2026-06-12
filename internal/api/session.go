package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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
	BaseURL      string
	Auth         string
	AuthType     string
	Headers      map[string]string
	Spec         string
	jqFilter     string        // optional jq filter for response formatting
	sessionStore *SessionStore // persistent cookie/auth store
	httpClient   *http.Client  // session-aware HTTP client (with cookie jar)
}

type SessionOption func(*Session)

// WithJqFilter sets a jq filter for the session.
func WithJqFilter(filter string) SessionOption {
	return func(s *Session) {
		s.jqFilter = filter
	}
}

// WithSessionStore enables HTTP session persistence for the session.
func WithSessionStore(store *SessionStore) SessionOption {
	return func(s *Session) {
		s.sessionStore = store
		if store != nil {
			s.httpClient = &http.Client{
				Timeout: httpRequestTimeout,
				Jar:     store.CookieJar(),
				Transport: &http.Transport{
					TLSHandshakeTimeout:   10 * time.Second,
					ResponseHeaderTimeout: 10 * time.Second,
				},
			}
		}
	}
}

func NewSession(cfg *config.API, opts ...SessionOption) (*Session, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no [api] section in tinker.toml")
	}
	if cfg.ResolvedBaseURL == "" {
		return nil, fmt.Errorf("api base_url is empty")
	}

	s := &Session{
		BaseURL:  normalizeBaseURL(cfg.ResolvedBaseURL),
		Auth:     cfg.ResolvedAuth,
		AuthType: cfg.AuthType,
		Headers:  cfg.Headers,
		Spec:     cfg.Spec,
	}

	for _, opt := range opts {
		opt(s)
	}

	// If session store has persisted auth, restore it
	if s.sessionStore != nil {
		if token, authType := s.sessionStore.GetAuth(); token != "" && s.Auth == "" {
			s.Auth = token
			s.AuthType = authType
		}
		// Merge persisted headers
		for k, v := range s.sessionStore.GetHeaders() {
			if _, exists := s.Headers[k]; !exists {
				if s.Headers == nil {
					s.Headers = make(map[string]string)
				}
				s.Headers[k] = v
			}
		}
	}

	return s, nil
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

// client returns the session-aware HTTP client (with cookie jar) if available,
// otherwise falls back to the default global client.
func (s *Session) client() *http.Client {
	if s.httpClient != nil {
		return s.httpClient
	}
	return httpClient
}
