package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (s *Session) Request(method, path, body string, extra map[string]string) (string, error) {
	var bodyR io.Reader
	if body != "" {
		bodyR = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, s.buildURL(path), bodyR)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	for k, v := range s.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range s.authHeaders() {
		req.Header.Set(k, v)
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
	for k, v := range resp.Header {
		fmt.Fprintf(&buf, "%s: %s\n", k, strings.Join(v, ", "))
	}
	buf.WriteString("\n")
	buf.Write(respBody)

	return buf.String(), nil
}

func (s *Session) Get(path string) (string, error)        { return s.Request("GET", path, "", nil) }
func (s *Session) Post(path, body string) (string, error) { return s.Request("POST", path, body, nil) }
func (s *Session) Put(path, body string) (string, error)  { return s.Request("PUT", path, body, nil) }
func (s *Session) Delete(path string) (string, error)     { return s.Request("DELETE", path, "", nil) }
