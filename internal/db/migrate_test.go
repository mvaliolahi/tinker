package db

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestParseMigrations(t *testing.T) {
	// Helper to create a temp directory with migration files
	createMigDir := func(t *testing.T, files map[string]string) string {
		t.Helper()
		dir := t.TempDir()
		for name, content := range files {
			fullPath := filepath.Join(dir, name)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("creating directory: %v", err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("writing file %s: %v", name, err)
			}
		}
		return dir
	}

	t.Run("empty directory returns empty", func(t *testing.T) {
		dir := t.TempDir()
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 0 {
			t.Errorf("expected 0 migrations, got %d", len(migs))
		}
	})

	t.Run("non-existent directory returns error", func(t *testing.T) {
		_, err := parseMigrations("/nonexistent/path/migrations")
		if err == nil {
			t.Error("expected error for non-existent directory, got nil")
		}
	})

	t.Run("single up migration", func(t *testing.T) {
		dir := createMigDir(t, map[string]string{
			"001_create_users.up.sql": "CREATE TABLE users (id INT);",
		})
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migs))
		}
		m := migs[0]
		if m.Version != "001" {
			t.Errorf("Version = %q, want %q", m.Version, "001")
		}
		if m.Name != "create_users" {
			t.Errorf("Name = %q, want %q", m.Name, "create_users")
		}
		if m.UpFile == "" {
			t.Error("UpFile should not be empty")
		}
		if m.DownFile != "" {
			t.Errorf("DownFile should be empty, got %q", m.DownFile)
		}
	})

	t.Run("up and down migrations", func(t *testing.T) {
		dir := createMigDir(t, map[string]string{
			"001_create_users.up.sql":   "CREATE TABLE users (id INT);",
			"001_create_users.down.sql": "DROP TABLE users;",
		})
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migs))
		}
		m := migs[0]
		if m.UpFile == "" {
			t.Error("UpFile should not be empty")
		}
		if m.DownFile == "" {
			t.Error("DownFile should not be empty")
		}
	})

	t.Run("multiple migrations sorted by version", func(t *testing.T) {
		dir := createMigDir(t, map[string]string{
			"003_add_emails.up.sql":   "ALTER TABLE users ADD COLUMN email TEXT;",
			"003_add_emails.down.sql": "ALTER TABLE users DROP COLUMN email;",
			"001_create_users.up.sql": "CREATE TABLE users (id INT);",
			"002_add_posts.up.sql":    "CREATE TABLE posts (id INT);",
		})
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 3 {
			t.Fatalf("expected 3 migrations, got %d", len(migs))
		}
		// Should be sorted by version
		wantVersions := []string{"001", "002", "003"}
		for i, want := range wantVersions {
			if migs[i].Version != want {
				t.Errorf("migs[%d].Version = %q, want %q", i, migs[i].Version, want)
			}
		}
	})

	t.Run("down-only migration is excluded", func(t *testing.T) {
		dir := createMigDir(t, map[string]string{
			"001_create_users.down.sql": "DROP TABLE users;",
		})
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 0 {
			t.Errorf("expected 0 migrations (no up file), got %d", len(migs))
		}
	})

	t.Run("files without underscore prefix are skipped", func(t *testing.T) {
		dir := createMigDir(t, map[string]string{
			"random.sql": "SELECT 1;",
		})
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 0 {
			t.Errorf("expected 0 migrations, got %d", len(migs))
		}
	})

	t.Run("non-sql files are ignored", func(t *testing.T) {
		dir := createMigDir(t, map[string]string{
			"001_create_users.up.sql": "CREATE TABLE users (id INT);",
			"001_create_users.up.txt": "not a migration",
		})
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 1 {
			t.Errorf("expected 1 migration, got %d", len(migs))
		}
	})

	t.Run("subdirectories are skipped", func(t *testing.T) {
		dir := t.TempDir()
		subdir := filepath.Join(dir, "subdir")
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatal(err)
		}
		// This file is inside a subdirectory — ReadDir won't recurse
		if err := os.WriteFile(filepath.Join(subdir, "001_sub.up.sql"), []byte("SELECT 1"), 0644); err != nil {
			t.Fatal(err)
		}
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 0 {
			t.Errorf("expected 0 migrations (subdir not read), got %d", len(migs))
		}
	})

	t.Run("migration file paths are absolute", func(t *testing.T) {
		dir := createMigDir(t, map[string]string{
			"001_create_users.up.sql": "CREATE TABLE users (id INT);",
		})
		migs, err := parseMigrations(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(migs) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migs))
		}
		if !filepath.IsAbs(migs[0].UpFile) {
			t.Errorf("UpFile = %q, expected absolute path", migs[0].UpFile)
		}
	})
}

func TestMigrationStruct(t *testing.T) {
	m := Migration{
		Version:  "001",
		Name:     "create_users",
		UpFile:   "/migrations/001_create_users.up.sql",
		DownFile: "/migrations/001_create_users.down.sql",
	}

	if m.Version != "001" {
		t.Errorf("Version = %q, want %q", m.Version, "001")
	}
	if m.Name != "create_users" {
		t.Errorf("Name = %q, want %q", m.Name, "create_users")
	}
	if m.UpFile == "" {
		t.Error("UpFile should not be empty")
	}
	if m.DownFile == "" {
		t.Error("DownFile should not be empty")
	}
}

func TestMigrationSorting(t *testing.T) {
	// Verify that the sorting logic used in parseMigrations produces
	// the correct order when versions are string-sorted.
	versions := []string{"003", "001", "010", "002"}
	sort.Strings(versions)

	want := []string{"001", "002", "003", "010"}
	for i, v := range versions {
		if v != want[i] {
			t.Errorf("sorted[%d] = %q, want %q", i, v, want[i])
		}
	}
}
