package detect

import (
	"os"
	"path/filepath"
	"strings"
)

type LogResult struct {
	Files []string
}

func (d *Detector) detectLog() *LogResult {
	result := ScanLogFiles(d.dir)
	if result == nil {
		return nil
	}
	return &LogResult{Files: result}
}

// skipDirs are directories to skip during recursive log file scanning.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".svn":         true,
	".hg":          true,
	"__pycache__":  true,
	".cache":       true,
	".next":        true,
	".nuxt":        true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".terraform":   true,
	"coverage":     true,
	".tox":         true,
	".venv":        true,
	"venv":         true,
}

// ScanLogFiles recursively searches for *.log files in a project directory.
// It skips common non-project directories like .git, node_modules, vendor, etc.
// Returns nil if no log files are found.
func ScanLogFiles(dir string) []string {
	var files []string

	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			// Skip hidden dirs and known non-project dirs
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a log file
		if isLogFile(d.Name()) {
			// Verify it's a regular file (not a directory named *.log)
			info, err := d.Info()
			if err != nil || info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				rel = path
			}
			files = append(files, rel)
		}

		return nil
	})

	if len(files) == 0 {
		return nil
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := files[:0]
	for _, f := range files {
		normalized := strings.TrimSpace(f)
		if !seen[normalized] {
			seen[normalized] = true
			unique = append(unique, normalized)
		}
	}

	return unique
}

// isLogFile checks if a filename looks like a log file.
func isLogFile(name string) bool {
	lower := strings.ToLower(name)

	// Match *.log and *.log.* (e.g., app.log, app.log.1, error.log.2024-01-15)
	if strings.HasSuffix(lower, ".log") {
		return true
	}
	// Match rotated/compressed logs: app.log.gz, error.log.old, access.log.1
	if idx := strings.Index(lower, ".log."); idx > 0 {
		return true
	}

	return false
}
