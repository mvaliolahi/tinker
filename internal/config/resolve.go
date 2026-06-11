package config

import (
	"fmt"
	"os"
	"strings"
)

func ResolveEnv(val string) (string, error) {
	if !strings.HasPrefix(val, "env:") {
		return val, nil
	}

	name := strings.TrimPrefix(val, "env:")
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("environment variable %s is not set", name)
	}
	return v, nil
}

func (c *Config) Resolve() error {
	if c.Database != nil {
		url, err := ResolveEnv(c.Database.Source)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		c.Database.URL = url
	}

	if c.API != nil {
		base, err := ResolveEnv(c.API.BaseURL)
		if err != nil {
			return fmt.Errorf("api.base_url: %w", err)
		}
		c.API.ResolvedBaseURL = base

		if c.API.Auth != "" {
			auth, err := ResolveEnv(c.API.Auth)
			if err != nil {
				return fmt.Errorf("api.auth: %w", err)
			}
			c.API.ResolvedAuth = auth
		}
	}

	if c.GRPC != nil {
		addr, err := ResolveEnv(c.GRPC.Addr)
		if err != nil {
			return fmt.Errorf("grpc.addr: %w", err)
		}
		c.GRPC.ResolvedAddr = addr
	}

	return nil
}
