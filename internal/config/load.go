package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mvaliolahi/tinker/internal/env"
	"github.com/pelletier/go-toml/v2"
)

// Load reads and parses the tinker.toml configuration file.
// Environment overrides are NOT applied — use LoadWithEnv for that.
func Load(dir string) (*Config, error) {
	return LoadWithEnv(dir, "")
}

// LoadWithEnv reads and parses the tinker.toml configuration file,
// then applies overrides for the specified environment (e.g., "staging", "production").
// An empty envName means use the base configuration without overrides.
func LoadWithEnv(dir string, envName string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
	if err != nil {
		return nil, fmt.Errorf("reading tinker.toml: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing tinker.toml: %w", err)
	}

	// Parse .env files into a map (no process-level mutation)
	envVars := env.ParseFiles(dir)
	cfg.SetEnvVars(envVars)

	// Apply environment-specific overrides before resolving
	if envName != "" {
		cfg.ApplyEnv(envName)
	}

	if err := cfg.Resolve(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
