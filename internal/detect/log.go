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

// ScanLogFiles searches for *.log files in a project directory.
// Returns nil if no log files are found.
func ScanLogFiles(dir string) []string {
        var files []string

        // Common log directories and patterns to search
        searchDirs := []string{".", "logs", "log", "var", "var/log", "tmp"}
        patterns := []string{"*.log", "*.log.*"}

        for _, dirName := range searchDirs {
                fullDir := filepath.Join(dir, dirName)
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
                                rel, err := filepath.Rel(dir, match)
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
                fullPath := filepath.Join(dir, name)
                if _, err := os.Stat(fullPath); err == nil {
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
