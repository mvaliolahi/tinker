package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

var envFiles = []string{".env", ".env.local", ".env.example"}

func Load(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
	if err != nil {
		return nil, fmt.Errorf("reading tinker.toml: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing tinker.toml: %w", err)
	}

	for _, name := range envFiles {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			if err := LoadEnvFile(p); err != nil {
				return nil, err
			}
			break
		}
	}

	if err := cfg.Resolve(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
