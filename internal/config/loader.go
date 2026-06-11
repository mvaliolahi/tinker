package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Load reads and parses tinker.toml from the given directory.
func Load(projectDir string) (*Config, error) {
	configPath := fmt.Sprintf("%s/tinker.toml", projectDir)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading tinker.toml: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing tinker.toml: %w", err)
	}

	// Load .env file if it exists
	envPath := fmt.Sprintf("%s/.env", projectDir)
	if _, err := os.Stat(envPath); err == nil {
		if err := LoadEnvFile(envPath); err != nil {
			return nil, fmt.Errorf("loading .env: %w", err)
		}
	}

	// Resolve all env: references
	if err := cfg.Resolve(); err != nil {
		return nil, fmt.Errorf("resolving config: %w", err)
	}

	return &cfg, nil
}
