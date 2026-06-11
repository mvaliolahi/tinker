package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mvaliolahi/tinker/internal/ui"
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

	fmt.Fprintf(&buf, "%s %s\n", ui.Dim("HTTP/"+protoVersion(resp)), formatStatus(resp.StatusCode))

	for k, v := range resp.Header {
		fmt.Fprintf(&buf, "%s %s\n", ui.Dim(k+":"), ui.Dim(strings.Join(v, ", ")))
	}

	bodyStr := string(respBody)
	if strings.TrimSpace(bodyStr) != "" {
		buf.WriteString("\n")
		if json.Valid([]byte(bodyStr)) {
			buf.WriteString(colorizeJSON(bodyStr))
		} else {
			buf.WriteString(bodyStr)
		}
	}

	return buf.String(), nil
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

func colorizeJSON(raw string) string {
	var buf bytes.Buffer
	if json.Indent(&buf, []byte(raw), "", "  ") != nil {
		return raw
	}
	indented := buf.String()

	// Tokenize: extract strings and non-strings, then colorize
	var out bytes.Buffer
	i := 0
	expectKey := true

	for i < len(indented) {
		ch := indented[i]

		if ch == '"' {
			// Read the full string (including quotes)
			end := findStringEnd(indented, i)
			token := indented[i:end]

			if expectKey {
				out.WriteString(ui.Bold(token))
			} else {
				out.WriteString(ui.Accent(token))
			}

			i = end
			expectKey = false
			continue
		}

		switch ch {
		case ':':
			out.WriteString(ui.Dim(": "))
			i += 2 // skip colon and space
			expectKey = false
		case ',':
			out.WriteString(ui.Dim(","))
			i++
			expectKey = true
		case '{', '[':
			out.WriteByte(ch)
			i++
			expectKey = true
		case '}', ']':
			out.WriteByte(ch)
			i++
		default:
			out.WriteByte(ch)
			i++
		}
	}

	return out.String()
}

func findStringEnd(s string, start int) int {
	i := start + 1 // skip opening quote
	for i < len(s) {
		if s[i] == '\\' {
			i += 2
			continue
		}
		if s[i] == '"' {
			return i + 1
		}
		i++
	}
	return len(s)
}

func (s *Session) Get(path string) (string, error)        { return s.Request("GET", path, "", nil) }
func (s *Session) Post(path, body string) (string, error) { return s.Request("POST", path, body, nil) }
func (s *Session) Put(path, body string) (string, error)  { return s.Request("PUT", path, body, nil) }
func (s *Session) Delete(path string) (string, error)     { return s.Request("DELETE", path, "", nil) }
