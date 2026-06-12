package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/tidwall/gjson"
)

const httpRequestTimeout = 30 * time.Second

var httpClient = &http.Client{
	Timeout: httpRequestTimeout,
	Transport: &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	},
}

func (s *Session) Request(method, path, body string, extra map[string]string) (string, error) {
	var bodyR io.Reader
	if body != "" {
		bodyR = strings.NewReader(body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, s.buildURL(path), bodyR)
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

	resp, err := s.client().Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("request timed out after %s", httpRequestTimeout)
		}
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Persist cookies and auth state from the response
	if s.sessionStore != nil {
		s.sessionStore.UpdateFromResponse(resp, s.BaseURL)
	}

	// Try gjson filtering first (native Go, no external dependency)
	if s.jqFilter != "" {
		if filtered := s.gjsonFilter(respBody); filtered != "" {
			respBody = []byte(filtered)
		} else {
			// gjson didn't match — try jq CLI as fallback
			if filtered, err := s.tryJqFilter(respBody); err == nil && filtered != "" {
				respBody = []byte(filtered)
			}
		}
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s %s\n", ui.Dim("HTTP/"+protoVersion(resp)), formatStatus(resp.StatusCode))

	for k, v := range resp.Header {
		fmt.Fprintf(&buf, "%s %s\n", ui.Dim(k+":"), ui.Dim(strings.Join(v, ", ")))
	}

	bodyStr := string(respBody)
	if strings.TrimSpace(bodyStr) != "" {
		buf.WriteString("\n")
		if json.Valid([]byte(bodyStr)) {
			buf.WriteString(highlightJSON(bodyStr))
		} else {
			buf.WriteString(bodyStr)
		}
	}

	return buf.String(), nil
}

// gjsonFilter applies a gjson path query to a JSON response body.
// gjson paths use dot notation (e.g., "data.users.0.name") which covers
// most common filtering needs without requiring the full jq language.
// Returns empty string if the path doesn't match or the body isn't JSON.
func (s *Session) gjsonFilter(body []byte) string {
	if !json.Valid(body) {
		return ""
	}

	result := gjson.GetBytes(body, s.jqFilter)
	if !result.Exists() {
		return ""
	}

	// Format the result based on its type
	switch result.Type {
	case gjson.String:
		return result.String()
	case gjson.Number:
		return result.String()
	case gjson.True, gjson.False:
		return result.String()
	case gjson.Null:
		return "null"
	case gjson.JSON:
		// Re-format the extracted JSON with indentation
		var pretty bytes.Buffer
		if json.Indent(&pretty, []byte(result.Raw), "", "  ") == nil {
			return pretty.String()
		}
		return result.Raw
	default:
		return result.Raw
	}
}

// tryJqFilter attempts to pipe the response through jq if a filter is set and jq is available.
// This is the fallback for complex jq expressions that gjson can't handle.
func (s *Session) tryJqFilter(body []byte) (string, error) {
	if s.jqFilter == "" {
		return "", nil
	}

	jqPath, err := exec.LookPath("jq")
	if err != nil {
		return "", fmt.Errorf("jq not found")
	}

	cmd := exec.Command(jqPath, s.jqFilter)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		_, _ = stdin.Write(body)
		_ = stdin.Close()
	}()

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func protoVersion(resp *http.Response) string {
	return fmt.Sprintf("%d.%d", resp.ProtoMajor, resp.ProtoMinor)
}

func formatStatus(code int) string {
	text := http.StatusText(code)
	if text == "" {
		text = "Unknown"
	}
	label := fmt.Sprintf("%d %s", code, text)

	switch {
	case code >= 200 && code < 300:
		return ui.Success(label)
	case code >= 300 && code < 400:
		return ui.Accent(label)
	case code >= 400 && code < 500:
		return ui.Warning(label)
	default:
		return ui.Error(label)
	}
}

// highlightJSON formats and syntax-highlights a JSON string using chroma.
func highlightJSON(raw string) string {
	// First indent the JSON
	var buf bytes.Buffer
	if json.Indent(&buf, []byte(raw), "", "  ") != nil {
		return raw
	}
	indented := buf.String()

	// Apply chroma syntax highlighting
	return ui.HighlightJSON(indented)
}
