package api

import "testing"

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://localhost:8080", "http://localhost:8080"},
		{"https://api.example.com", "https://api.example.com"},
		{"localhost:3000", "http://localhost:3000"},
		{":8080", "http://localhost:8080"},
		{"8080", "http://localhost:8080"},
		{"api.example.com/v1", "http://api.example.com/v1"},
		{"http://localhost:8080/", "http://localhost:8080"},
	}
	for _, tt := range tests {
		got := normalizeBaseURL(tt.input)
		if got != tt.want {
			t.Errorf("normalizeBaseURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsPortOnly(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"8080", true},
		{"3000", true},
		{"", false},
		{"localhost", false},
		{"8080a", false},
	}
	for _, tt := range tests {
		got := isPortOnly(tt.input)
		if got != tt.want {
			t.Errorf("isPortOnly(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
