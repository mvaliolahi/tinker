package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := `# Comment
DB_HOST=localhost
DB_PORT=5432
DB_URL=postgres://user:pass@localhost:5432/mydb
EMPTY_VAR=
QUOTED_VAR="hello world"
SINGLE_QUOTED='single'
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	env := ParseFile(path)

	if env["DB_HOST"] != "localhost" {
		t.Errorf("DB_HOST = %q, want %q", env["DB_HOST"], "localhost")
	}
	if env["DB_PORT"] != "5432" {
		t.Errorf("DB_PORT = %q, want %q", env["DB_PORT"], "5432")
	}
	if env["DB_URL"] != "postgres://user:pass@localhost:5432/mydb" {
		t.Errorf("DB_URL = %q, want %q", env["DB_URL"], "postgres://user:pass@localhost:5432/mydb")
	}
	if env["QUOTED_VAR"] != "hello world" {
		t.Errorf("QUOTED_VAR = %q, want %q", env["QUOTED_VAR"], "hello world")
	}
	if env["SINGLE_QUOTED"] != "single" {
		t.Errorf("SINGLE_QUOTED = %q, want %q", env["SINGLE_QUOTED"], "single")
	}
	if _, ok := env["EMPTY_VAR"]; ok {
		t.Error("EMPTY_VAR should not be present (empty value)")
	}
	if _, ok := env["# Comment"]; ok {
		t.Error("comments should not be parsed")
	}
}

func TestParseFileNonExistent(t *testing.T) {
	env := ParseFile("/nonexistent/.env")
	if len(env) != 0 {
		t.Errorf("expected empty map for nonexistent file, got %v", env)
	}
}

func TestParseFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("KEY1=val1\nKEY2=val2"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("KEY2=overridden\nKEY3=val3"), 0644); err != nil {
		t.Fatal(err)
	}

	env := ParseFiles(dir)

	if env["KEY1"] != "val1" {
		t.Errorf("KEY1 = %q, want %q", env["KEY1"], "val1")
	}
	// .env is read first, .env.local should not override (first-wins)
	if env["KEY2"] != "val2" {
		t.Errorf("KEY2 = %q, want %q (first file wins)", env["KEY2"], "val2")
	}
	if env["KEY3"] != "val3" {
		t.Errorf("KEY3 = %q, want %q", env["KEY3"], "val3")
	}
}

func TestResolve(t *testing.T) {
	envVars := map[string]string{
		"DATABASE_URL": "postgres://localhost:5432/db",
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"plain value", "hello", "hello", false},
		{"env var found", "env:DATABASE_URL", "postgres://localhost:5432/db", false},
		{"env var not found", "env:MISSING_VAR", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(envVars, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveFallbackToOSEnv(t *testing.T) {
	os.Setenv("TINKER_TEST_VAR", "from_os")
	defer os.Unsetenv("TINKER_TEST_VAR")

	envVars := map[string]string{}
	got, err := Resolve(envVars, "env:TINKER_TEST_VAR")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}
	if got != "from_os" {
		t.Errorf("Resolve() = %q, want %q", got, "from_os")
	}
}
