package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindProjectRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "tinker.toml")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("tinker.toml not found — run 'tinker init' first")
		}
		dir = parent
	}
}
