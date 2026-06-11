package config

import (
	"fmt"
	"os"
	"strings"
)

func LoadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading .env: %w", err)
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

		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}

	return nil
}

func ParseEnvFile(path string) map[string]string {
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
		env[k] = v
	}

	return env
}
