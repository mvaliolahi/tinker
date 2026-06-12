package config

import "testing"

func TestValidateEmpty(t *testing.T) {
	cfg := &Config{}
	issues := cfg.Validate()
	if len(issues) == 0 {
		t.Error("expected issues for empty config")
	}
	if issues[0] != "no services configured (database, api, grpc, or log)" {
		t.Errorf("unexpected issue: %q", issues[0])
	}
}

func TestValidateDatabaseEmptySource(t *testing.T) {
	cfg := &Config{
		Database: &Database{Source: "", Type: "postgres"},
	}
	issues := cfg.Validate()
	found := false
	for _, issue := range issues {
		if issue == "[database] source is empty" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'source is empty' issue")
	}
}

func TestValidateAPIEmptyBaseURL(t *testing.T) {
	cfg := &Config{
		API: &API{BaseURL: ""},
	}
	issues := cfg.Validate()
	found := false
	for _, issue := range issues {
		if issue == "[api] base_url is empty" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'base_url is empty' issue")
	}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := &Config{
		Database: &Database{Source: "env:DATABASE_URL", Type: "postgres", URL: "postgres://localhost/db"},
		API:      &API{BaseURL: "env:API_URL", ResolvedBaseURL: "http://localhost:8080"},
	}
	issues := cfg.Validate()
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}
