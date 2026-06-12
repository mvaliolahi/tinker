package config

import (
	"fmt"

	"github.com/mvaliolahi/tinker/internal/env"
)

func (c *Config) Resolve() error {
	ev := c.GetEnvVars()

	if c.Database != nil {
		url, err := env.Resolve(ev, c.Database.Source)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		c.Database.URL = url
	}

	if c.API != nil {
		base, err := env.Resolve(ev, c.API.BaseURL)
		if err != nil {
			return fmt.Errorf("api.base_url: %w", err)
		}
		c.API.ResolvedBaseURL = base

		if c.API.Auth != "" {
			auth, err := env.Resolve(ev, c.API.Auth)
			if err != nil {
				return fmt.Errorf("api.auth: %w", err)
			}
			c.API.ResolvedAuth = auth
		}
	}

	if c.GRPC != nil {
		addr, err := env.Resolve(ev, c.GRPC.Addr)
		if err != nil {
			return fmt.Errorf("grpc.addr: %w", err)
		}
		c.GRPC.ResolvedAddr = addr
	}

	return nil
}

// Validate checks the configuration for common issues and returns all problems found.
func (c *Config) Validate() []string {
	var issues []string

	if c.Database == nil && c.API == nil && c.GRPC == nil && c.Log == nil {
		issues = append(issues, "no services configured (database, api, grpc, or log)")
	}

	if c.Database != nil {
		if c.Database.Source == "" {
			issues = append(issues, "[database] source is empty")
		}
		if c.Database.Type == "" {
			issues = append(issues, "[database] type is empty")
		}
		if c.Database.URL == "" {
			issues = append(issues, "[database] URL could not be resolved (check source and env vars)")
		}
	}

	if c.API != nil {
		if c.API.BaseURL == "" {
			issues = append(issues, "[api] base_url is empty")
		}
		if c.API.ResolvedBaseURL == "" {
			issues = append(issues, "[api] base_url could not be resolved (check env vars)")
		}
	}

	if c.GRPC != nil {
		if c.GRPC.Addr == "" && c.GRPC.ProtoDir == "" {
			issues = append(issues, "[grpc] both addr and proto_dir are empty")
		}
	}

	if c.Log != nil && len(c.Log.Files) == 0 {
		issues = append(issues, "[log] files list is empty")
	}

	return issues
}
