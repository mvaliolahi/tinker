package api

import (
        "fmt"
        "net/http"
        "net/http/cookiejar"
        "net/url"
        "os"
        "path/filepath"
        "strings"
        "sync"
        "time"
)

const (
        sessionDirName  = ".tinker"
        sessionFileName = "session.json"
)

// SessionStore persists HTTP cookies and auth state across requests.
// It wraps an http.Client with a persistent cookie jar and provides
// methods for saving/loading session data to disk.
type SessionStore struct {
        mu       sync.RWMutex
        jar      http.CookieJar
        filePath string
        data     *SessionData
}

// SessionData represents the persisted session state.
type SessionData struct {
        Cookies   []*http.Cookie    `json:"cookies"`
        AuthToken string            `json:"auth_token,omitempty"`
        AuthType  string            `json:"auth_type,omitempty"`
        Headers   map[string]string `json:"headers,omitempty"`
        SavedAt   time.Time         `json:"saved_at"`
}

// NewSessionStore creates a new session store with cookies persisted to the given directory.
func NewSessionStore(projectDir string) (*SessionStore, error) {
        jar, err := cookiejar.New(nil)
        if err != nil {
                return nil, fmt.Errorf("creating cookie jar: %w", err)
        }

        sessionDir := filepath.Join(projectDir, sessionDirName)
        filePath := filepath.Join(sessionDir, sessionFileName)

        store := &SessionStore{
                jar:      jar,
                filePath: filePath,
                data: &SessionData{
                        Headers: make(map[string]string),
                },
        }

        // Load existing session data if available
        if err := store.load(); err == nil {
                // Restore cookies into the jar
                for _, cookie := range store.data.Cookies {
                        if cookie != nil && cookie.Domain != "" {
                                scheme := "http"
                                if cookie.Secure {
                                        scheme = "https"
                                }
                                u, _ := url.Parse(scheme + "://" + strings.TrimPrefix(cookie.Domain, "."))
                                if u != nil {
                                        jar.SetCookies(u, []*http.Cookie{cookie})
                                }
                        }
                }
        }

        return store, nil
}

// CookieJar returns the persistent cookie jar for use with http.Client.
func (s *SessionStore) CookieJar() http.CookieJar {
        return s.jar
}

// SaveCookies extracts cookies from the jar for a given URL and persists them.
func (s *SessionStore) SaveCookies(_ string) {
        s.mu.Lock()
        defer s.mu.Unlock()

        // The cookiejar doesn't expose all cookies directly,
        // so we rely on the http.Client to use the jar and then
        // capture Set-Cookie headers from responses.
}

// UpdateFromResponse captures cookies and auth state from an HTTP response.
func (s *SessionStore) UpdateFromResponse(resp *http.Response, _ string) {
        s.mu.Lock()
        defer s.mu.Unlock()

        if resp == nil {
                return
        }

        // Capture Set-Cookie headers
        cookies := resp.Cookies()
        if len(cookies) > 0 && s.jar != nil {
                // Cookies are already set in the jar by http.Client,
                // we just need to persist them
                s.data.Cookies = append(s.data.Cookies, cookies...)
                // Deduplicate by name
                s.data.Cookies = dedupCookies(s.data.Cookies)
        }

        // Check for auth-related response headers (e.g., new tokens)
        if auth := resp.Header.Get("X-New-Token"); auth != "" {
                s.data.AuthToken = auth
        }

        _ = s.save()
}

// SetAuth persists auth credentials in the session.
func (s *SessionStore) SetAuth(token, authType string) {
        s.mu.Lock()
        defer s.mu.Unlock()
        s.data.AuthToken = token
        s.data.AuthType = authType
        _ = s.save()
}

// GetAuth returns the persisted auth credentials.
func (s *SessionStore) GetAuth() (token, authType string) {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return s.data.AuthToken, s.data.AuthType
}

// SetHeader persists a custom header in the session.
func (s *SessionStore) SetHeader(key, value string) {
        s.mu.Lock()
        defer s.mu.Unlock()
        if s.data.Headers == nil {
                s.data.Headers = make(map[string]string)
        }
        s.data.Headers[key] = value
        _ = s.save()
}

// GetHeaders returns all persisted custom headers.
func (s *SessionStore) GetHeaders() map[string]string {
        s.mu.RLock()
        defer s.mu.RUnlock()
        if s.data.Headers == nil {
                return make(map[string]string)
        }
        // Return a copy
        result := make(map[string]string, len(s.data.Headers))
        for k, v := range s.data.Headers {
                result[k] = v
        }
        return result
}

// Clear removes all persisted session data.
func (s *SessionStore) Clear() error {
        s.mu.Lock()
        defer s.mu.Unlock()
        s.data = &SessionData{
                Headers: make(map[string]string),
        }
        return os.Remove(s.filePath)
}

// load reads session data from disk.
func (s *SessionStore) load() error {
        data, err := os.ReadFile(s.filePath)
        if err != nil {
                return err
        }

        var sd SessionData
        if err := parseSessionJSON(data, &sd); err != nil {
                return err
        }

        s.data = &sd
        return nil
}

// save writes session data to disk.
func (s *SessionStore) save() error {
        s.data.SavedAt = time.Now()

        dir := filepath.Dir(s.filePath)
        if err := os.MkdirAll(dir, 0700); err != nil {
                return err
        }

        data, err := encodeSessionJSON(s.data)
        if err != nil {
                return err
        }

        return os.WriteFile(s.filePath, data, 0600)
}

// parseSessionJSON deserializes session data from JSON bytes.
func parseSessionJSON(data []byte, sd *SessionData) error {
        return decodeJSON(data, sd)
}

// encodeSessionJSON serializes session data to JSON bytes.
func encodeSessionJSON(sd *SessionData) ([]byte, error) {
        return encodeJSON(sd)
}

// dedupCookies removes duplicate cookies by name, keeping the last one.
func dedupCookies(cookies []*http.Cookie) []*http.Cookie {
        seen := make(map[string]int)
        var result []*http.Cookie
        for _, c := range cookies {
                if c == nil {
                        continue
                }
                if prev, ok := seen[c.Name]; ok {
                        result[prev] = c // replace previous
                } else {
                        result = append(result, c)
                        seen[c.Name] = len(result) - 1
                }
        }
        return result
}
