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
	var files []string

	// Common log directories and patterns to search
	searchDirs := []string{".", "logs", "log", "var", "var/log", "tmp"}
	patterns := []string{"*.log", "*.log.*"}

	for _, dir := range searchDirs {
		fullDir := filepath.Join(d.dir, dir)
		info, err := os.Stat(fullDir)
		if err != nil || !info.IsDir() {
			continue
		}

		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(fullDir, pattern))
			if err != nil {
				continue
			}
			for _, match := range matches {
				rel, err := filepath.Rel(d.dir, match)
				if err != nil {
					rel = match
				}
				files = append(files, rel)
			}
		}
	}

	// Also check common log file names at project root
	rootLogs := []string{"app.log", "server.log", "error.log", "access.log", "debug.log", "application.log", "out.log"}
	for _, name := range rootLogs {
		fullPath := filepath.Join(d.dir, name)
		if _, err := os.Stat(fullPath); err == nil {
			// Avoid duplicate
			found := false
			for _, f := range files {
				if f == name {
					found = true
					break
				}
			}
			if !found {
				files = append(files, name)
			}
		}
	}

	// Check environment variables for log file paths
	envKeys := []string{"LOG_FILE", "LOG_PATH", "ACCESS_LOG", "ERROR_LOG", "APP_LOG"}
	for _, key := range envKeys {
		if v := d.getEnv(key); v != "" {
			// If it's a relative path or the file exists, add it
			if !filepath.IsAbs(v) {
				if _, err := os.Stat(filepath.Join(d.dir, v)); err == nil {
					files = append(files, v)
				}
			}
		}
	}

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

	return &LogResult{Files: unique}
}
