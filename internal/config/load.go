package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mvaliolahi/tinker/internal/env"
	"github.com/pelletier/go-toml/v2"
)

func Load(dir string) (*Config, error) {
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

	if err := cfg.Resolve(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
