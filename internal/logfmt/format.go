package logfmt

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mvaliolahi/tinker/internal/ui"
)

// FormatLine detects JSON log lines and formats them; falls back to text colorization.
func FormatLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	if line[0] == '{' {
		var obj map[string]interface{}
		if json.Unmarshal([]byte(line), &obj) == nil {
			return FormatJSON(obj)
		}
	}
	return ColorizeText(line)
}

// FormatJSON renders a structured JSON log line as a clean, colorized single line.
func FormatJSON(obj map[string]interface{}) string {
	var sb strings.Builder

	ts := FieldStr(obj, "ts", "time", "t", "timestamp", "datetime")
	if ts != "" {
		sb.WriteString(ui.Dim(FormatTimestamp(ts)))
		sb.WriteString(" ")
	}

	level := FieldStr(obj, "level", "lvl", "severity")
	if level != "" {
		sb.WriteString(FormatLevel(level))
		sb.WriteString(" ")
	}

	method := FieldStr(obj, "method")
	path := FieldStr(obj, "path", "url", "uri")
	status := FieldInt(obj, "status", "status_code", "statusCode")
	duration := FieldInt(obj, "duration", "dur", "latency")

	if method != "" || path != "" {
		if method != "" {
			sb.WriteString(ui.HTTPMethod(method))
			sb.WriteString(" ")
		}
		if path != "" {
			sb.WriteString(ui.Accent(path))
		}
		if status > 0 {
			sb.WriteString(" ")
			sb.WriteString(ui.Dim("→"))
			sb.WriteString(" ")
			sb.WriteString(FormatStatus(status))
		}
		if duration > 0 {
			sb.WriteString("  ")
			sb.WriteString(ui.Dim(FormatDuration(duration)))
		}
	} else {
		msg := FieldStr(obj, "msg", "message", "text")
		if msg != "" {
			sb.WriteString(msg)
		}
	}

	event := FieldStr(obj, "event")
	if event != "" && method == "" {
		sb.WriteString(" ")
		sb.WriteString(ui.Dim(event))
	}

	signal := FieldStr(obj, "signal")
	if signal != "" {
		sb.WriteString(" ")
		sb.WriteString(ui.Warning(signal))
	}

	shown := map[string]bool{
		"ts": true, "time": true, "t": true, "timestamp": true, "datetime": true,
		"level": true, "lvl": true, "severity": true,
		"msg": true, "message": true, "text": true,
		"method": true, "path": true, "url": true, "uri": true,
		"status": true, "status_code": true, "statusCode": true,
		"duration": true, "dur": true, "latency": true,
		"event": true, "signal": true,
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		if !shown[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if len(keys) > 0 {
		sb.WriteString("  ")
		for i, k := range keys {
			if i > 0 {
				sb.WriteString("  ")
			}
			v := obj[k]
			sb.WriteString(ui.Dim(k + "="))
			sb.WriteString(FormatValue(v))
		}
	}

	return sb.String()
}

// FormatLevel returns a colorized, left-aligned level string.
func FormatLevel(level string) string {
	upper := strings.ToUpper(level)
	padded := fmt.Sprintf("%-5s", upper)
	switch upper {
	case "ERROR", "FATAL", "PANIC", "CRITICAL":
		return ui.Error(padded)
	case "WARN", "WARNING":
		return ui.Warning(padded)
	case "INFO", "NOTICE":
		return ui.Success(padded)
	case "DEBUG", "TRACE":
		return ui.Dim(padded)
	default:
		return padded
	}
}

// FormatStatus returns a colorized HTTP status code.
func FormatStatus(code int64) string {
	s := fmt.Sprintf("%d", code)
	switch {
	case code >= 200 && code < 300:
		return ui.Success(s)
	case code >= 300 && code < 400:
		return ui.Accent(s)
	case code >= 400 && code < 500:
		return ui.Warning(s)
	default:
		return ui.Error(s)
	}
}

// FormatDuration formats nanoseconds as a human-readable duration.
func FormatDuration(ns int64) string {
	switch {
	case ns >= 1_000_000_000:
		return fmt.Sprintf("%.1fs", float64(ns)/1_000_000_000)
	case ns >= 1_000_000:
		return fmt.Sprintf("%.1fms", float64(ns)/1_000_000)
	case ns >= 1_000:
		return fmt.Sprintf("%.1fµs", float64(ns)/1_000)
	default:
		return fmt.Sprintf("%dns", ns)
	}
}

// FormatTimestamp converts an ISO timestamp to a friendlier format.
func FormatTimestamp(ts string) string {
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, ts); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}
	return ts
}

// FormatValue renders a JSON value as a short string for key=value display.
func FormatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		if val == "" {
			return ui.Dim(`""`)
		}
		if len(val) > 60 {
			return ui.Dim(val[:57] + "...")
		}
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		if val {
			return ui.Success("true")
		}
		return ui.Error("false")
	case nil:
		return ui.Dim("null")
	default:
		s := fmt.Sprintf("%v", val)
		if len(s) > 60 {
			s = s[:57] + "..."
		}
		return ui.Dim(s)
	}
}

// FieldStr extracts a string value from a map, trying multiple key names.
func FieldStr(obj map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

// FieldInt extracts an int64 value from a map, trying multiple key names.
func FieldInt(obj map[string]interface{}, keys ...string) int64 {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			switch n := v.(type) {
			case float64:
				return int64(n)
			case int64:
				return n
			case int:
				return int64(n)
			}
		}
	}
	return 0
}

// MatchLevel checks if a log line contains a specific log level.
func MatchLevel(line, level string) bool {
	upper := strings.ToUpper(level)
	lineUpper := strings.ToUpper(line)

	// Try JSON
	if len(line) > 0 && line[0] == '{' {
		var obj map[string]interface{}
		if json.Unmarshal([]byte(line), &obj) == nil {
			lineLevel := strings.ToUpper(FieldStr(obj, "level", "lvl", "severity"))
			return strings.Contains(lineLevel, upper)
		}
	}

	// Text patterns
	patterns := []string{
		"[" + upper + "]",
		upper + ":",
	}
	for _, p := range patterns {
		if strings.Contains(lineUpper, p) {
			return true
		}
	}
	return false
}

// ColorizeText handles non-JSON log lines (plain text).
func ColorizeText(line string) string {
	line = colorizeLevels(line)
	line = colorizeTimestamps(line)
	return line
}

func colorizeLevels(line string) string {
	replacements := []struct {
		pattern string
		style   func(string) string
	}{
		{"[ERROR]", ui.Error}, {"[FATAL]", ui.Error}, {"[PANIC]", ui.Error}, {"[CRITICAL]", ui.Error},
		{"[WARN]", ui.Warning}, {"[WARNING]", ui.Warning},
		{"[INFO]", ui.Success}, {"[NOTICE]", ui.Success},
		{"[DEBUG]", ui.Dim}, {"[TRACE]", ui.Dim},
		{"ERROR:", ui.Error}, {"FATAL:", ui.Error}, {"PANIC:", ui.Error},
		{"WARN:", ui.Warning}, {"WARNING:", ui.Warning},
		{"INFO:", ui.Success}, {"NOTICE:", ui.Success},
		{"DEBUG:", ui.Dim}, {"TRACE:", ui.Dim},
	}
	for _, r := range replacements {
		if strings.Contains(line, r.pattern) {
			line = strings.ReplaceAll(line, r.pattern, r.style(r.pattern))
		}
	}
	return line
}

func colorizeTimestamps(line string) string {
	if len(line) > 4 {
		if line[:4] >= "2000" && line[:4] <= "2099" && len(line) > 4 && (line[4] == '-' || line[4] == '/') {
			end := 25
			if len(line) < end {
				end = len(line)
			}
			ts := line[:end]
			rest := ""
			if len(line) > end {
				rest = line[end:]
			}
			return ui.Dim(ts) + rest
		}
	}
	return line
}
