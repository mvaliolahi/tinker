package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var envFiles = []string{".env", ".env.local", ".env.example", ".env.sample"}

// ParseFile reads a single .env file and returns key-value pairs.
func ParseFile(path string) map[string]string {
	env := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return env
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if v != "" {
			env[k] = v
		}
	}
	return env
}

// ParseFiles reads all .env files from a directory and merges them.
func ParseFiles(dir string) map[string]string {
	env := make(map[string]string)
	for _, name := range envFiles {
		for k, v := range ParseFile(filepath.Join(dir, name)) {
			if _, exists := env[k]; !exists {
				env[k] = v
			}
		}
	}
	return env
}

// Resolve resolves a "env:VAR_NAME" reference using the provided env map,
// falling back to os.Getenv if not found in the map.
func Resolve(envVars map[string]string, val string) (string, error) {
	if !strings.HasPrefix(val, "env:") {
		return val, nil
	}
	name := strings.TrimPrefix(val, "env:")
	if v, ok := envVars[name]; ok && v != "" {
		return v, nil
	}
	if v := os.Getenv(name); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("environment variable %s is not set", name)
}
