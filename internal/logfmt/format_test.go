package logfmt

import (
	"testing"
)

func TestFormatLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"error", "ERROR"},
		{"warn", "WARN"},
		{"info", "INFO"},
		{"debug", "DEBUG"},
		{"fatal", "FATAL"},
		{"trace", "TRACE"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatLevel(tt.input)
			// The output contains ANSI codes, so just check it contains the expected text
			if !contains(got, tt.want) {
				t.Errorf("FormatLevel(%q) = %q, want to contain %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		code int64
		want string
	}{
		{200, "200"},
		{301, "301"},
		{404, "404"},
		{500, "500"},
	}
	for _, tt := range tests {
		got := FormatStatus(tt.code)
		if !contains(got, tt.want) {
			t.Errorf("FormatStatus(%d) = %q, want to contain %q", tt.code, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ns   int64
		want string
	}{
		{500, "500ns"},
		{1500, "1.5µs"},
		{1_500_000, "1.5ms"},
		{1_500_000_000, "1.5s"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.ns)
		if got != tt.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.ns, got, tt.want)
		}
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-06-12T10:30:00Z", "2026-06-12 10:30:00"},
		{"2026-06-12 15:45:30", "2026-06-12 15:45:30"},
	}
	for _, tt := range tests {
		got := FormatTimestamp(tt.input)
		if got != tt.want {
			t.Errorf("FormatTimestamp(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFieldStr(t *testing.T) {
	obj := map[string]interface{}{
		"level": "info",
		"msg":   "hello",
	}
	if got := FieldStr(obj, "level"); got != "info" {
		t.Errorf("FieldStr(level) = %q, want %q", got, "info")
	}
	if got := FieldStr(obj, "missing", "msg"); got != "hello" {
		t.Errorf("FieldStr(missing, msg) = %q, want %q", got, "hello")
	}
	if got := FieldStr(obj, "nonexistent"); got != "" {
		t.Errorf("FieldStr(nonexistent) = %q, want empty", got)
	}
}

func TestFieldInt(t *testing.T) {
	obj := map[string]interface{}{
		"status":   float64(200),
		"duration": int64(1500),
		"count":    42,
	}
	if got := FieldInt(obj, "status"); got != 200 {
		t.Errorf("FieldInt(status) = %d, want 200", got)
	}
	if got := FieldInt(obj, "duration"); got != 1500 {
		t.Errorf("FieldInt(duration) = %d, want 1500", got)
	}
	if got := FieldInt(obj, "count"); got != 42 {
		t.Errorf("FieldInt(count) = %d, want 42", got)
	}
}

func TestMatchLevel(t *testing.T) {
	tests := []struct {
		line  string
		level string
		want  bool
	}{
		{`{"level":"error","msg":"fail"}`, "error", true},
		{`{"level":"info","msg":"ok"}`, "error", false},
		{`[ERROR] something failed`, "error", true},
		{`INFO: all good`, "info", true},
		{`plain text without level`, "error", false},
	}
	for _, tt := range tests {
		got := MatchLevel(tt.line, tt.level)
		if got != tt.want {
			t.Errorf("MatchLevel(%q, %q) = %v, want %v", tt.line, tt.level, got, tt.want)
		}
	}
}

func TestFormatLine(t *testing.T) {
	// Empty line
	if got := FormatLine(""); got != "" {
		t.Errorf("FormatLine('') = %q, want empty", got)
	}
	// Whitespace line
	if got := FormatLine("   "); got != "" {
		t.Errorf("FormatLine('   ') = %q, want empty", got)
	}
	// Text line
	got := FormatLine("[INFO] server started")
	if !contains(got, "INFO") {
		t.Errorf("FormatLine text = %q, want to contain INFO", got)
	}
}

func TestColorizeText(t *testing.T) {
	got := ColorizeText("[ERROR] something bad")
	if !contains(got, "ERROR") {
		t.Errorf("ColorizeText = %q, want to contain ERROR", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
