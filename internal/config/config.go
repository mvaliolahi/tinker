// Package config handles reading and parsing tinker.toml, .env, and related configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the full tinker configuration.
type Config struct {
	Database *DatabaseConfig `toml:"database"`
	API      *APIConfig      `toml:"api"`
	GRPC     *GRPCConfig     `toml:"grpc"`
	Commands map[string]string `toml:"commands"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	// Source specifies how to obtain the connection string.
	// Format: "env:VAR_NAME" to read from environment, or a direct DSN string.
	Source string `toml:"source"`
	// Type of database: postgres, mysql, sqlite3, etc.
	Type string `toml:"type"`
	// Driver overrides the default driver name derived from Type.
	Driver string `toml:"driver"`
	// URL is the resolved connection string after processing Source.
	URL string `toml:"-"`
}

// APIConfig holds HTTP API settings.
type APIConfig struct {
	// BaseURL source: "env:API_BASE_URL" or direct URL like "http://localhost:8080"
	BaseURL string `toml:"base_url"`
	// Spec path to OpenAPI/Swagger spec file (optional).
	Spec string `toml:"spec"`
	// Auth source for authentication header: "env:API_TOKEN" or direct value.
	Auth string `toml:"auth"`
	// AuthType determines the auth header format: "bearer", "basic", "api_key", or raw.
	AuthType string `toml:"auth_type"`
	// Headers contains additional default headers.
	Headers map[string]string `toml:"headers"`
	// ResolvedBaseURL is the resolved base URL after processing.
	ResolvedBaseURL string `toml:"-"`
	// ResolvedAuth is the resolved auth value after processing.
	ResolvedAuth string `toml:"-"`
}

// GRPCConfig holds gRPC service settings.
type GRPCConfig struct {
	// Addr source: "env:GRPC_ADDR" or direct address like "localhost:50051"
	Addr string `toml:"addr"`
	// ProtoDir path to directory containing .proto files.
	ProtoDir string `toml:"proto_dir"`
	// Reflection enables gRPC server reflection (grpcurl without proto files).
	Reflection bool `toml:"reflection"`
	// ResolvedAddr is the resolved address after processing.
	ResolvedAddr string `toml:"-"`
}

// Resolve processes all "env:" prefixed values and fills in resolved fields.
func (c *Config) Resolve() error {
	if c.Database != nil {
		url, err := resolveValue(c.Database.Source)
		if err != nil {
			return fmt.Errorf("resolving database source: %w", err)
		}
		c.Database.URL = url
	}

	if c.API != nil {
		baseURL, err := resolveValue(c.API.BaseURL)
		if err != nil {
			return fmt.Errorf("resolving api base_url: %w", err)
		}
		c.API.ResolvedBaseURL = baseURL

		if c.API.Auth != "" {
			auth, err := resolveValue(c.API.Auth)
			if err != nil {
				return fmt.Errorf("resolving api auth: %w", err)
			}
			c.API.ResolvedAuth = auth
		}
	}

	if c.GRPC != nil {
		addr, err := resolveValue(c.GRPC.Addr)
		if err != nil {
			return fmt.Errorf("resolving grpc addr: %w", err)
		}
		c.GRPC.ResolvedAddr = addr
	}

	return nil
}

// resolveValue checks if a value starts with "env:" and reads from the environment.
// Otherwise, returns the value as-is.
func resolveValue(val string) (string, error) {
	if strings.HasPrefix(val, "env:") {
		envVar := strings.TrimPrefix(val, "env:")
		envVal := os.Getenv(envVar)
		if envVal == "" {
			return "", fmt.Errorf("environment variable %s is not set", envVar)
		}
		return envVal, nil
	}
	return val, nil
}

// LoadEnvFile reads a .env file and sets environment variables.
// It does not override existing environment variables.
func LoadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading .env file: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip surrounding quotes
		value = strings.Trim(value, `"'`)

		// Don't override existing env vars
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

// FindProjectRoot walks up the directory tree looking for tinker.toml.
func FindProjectRoot(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "tinker.toml")
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("tinker.toml not found in any parent directory")
		}
		dir = parent
	}
}
