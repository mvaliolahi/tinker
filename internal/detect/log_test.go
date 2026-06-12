package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanLogFiles(t *testing.T) {
	dir := t.TempDir()

	// Create some log files
	os.MkdirAll(filepath.Join(dir, "logs"), 0755)
	os.WriteFile(filepath.Join(dir, "logs", "app.log"), []byte("log"), 0644)
	os.WriteFile(filepath.Join(dir, "error.log"), []byte("log"), 0644)

	// Create files that should be skipped
	os.MkdirAll(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, ".git", "debug.log"), []byte("log"), 0644)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg.log"), []byte("log"), 0644)

	found := ScanLogFiles(dir)
	if len(found) != 2 {
		t.Errorf("ScanLogFiles found %d files, want 2: %v", len(found), found)
	}

	hasAppLog := false
	hasErrorLog := false
	for _, f := range found {
		if f == filepath.Join("logs", "app.log") {
			hasAppLog = true
		}
		if f == "error.log" {
			hasErrorLog = true
		}
	}
	if !hasAppLog {
		t.Error("expected to find logs/app.log")
	}
	if !hasErrorLog {
		t.Error("expected to find error.log")
	}
}

func TestScanLogFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	found := ScanLogFiles(dir)
	if found != nil {
		t.Errorf("ScanLogFiles on empty dir = %v, want nil", found)
	}
}
