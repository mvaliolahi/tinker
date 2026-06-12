package config

type Config struct {
	Database *Database `toml:"database"`
	API      *API      `toml:"api"`
	GRPC     *GRPC     `toml:"grpc"`
	Log      *Log      `toml:"log"`
	Commands Commands  `toml:"commands"`
	Envs     Envs      `toml:"envs"`

	// envVars holds parsed .env values (not persisted to TOML)
	envVars map[string]string `toml:"-"`
}

type Database struct {
	Source string `toml:"source"`
	Type   string `toml:"type"`
	Driver string `toml:"driver"`
	URL    string `toml:"-"`
}

type API struct {
	BaseURL  string            `toml:"base_url"`
	Spec     string            `toml:"spec"`
	Auth     string            `toml:"auth"`
	AuthType string            `toml:"auth_type"`
	Headers  map[string]string `toml:"headers"`

	ResolvedBaseURL string `toml:"-"`
	ResolvedAuth    string `toml:"-"`
}

type GRPC struct {
	Addr       string `toml:"addr"`
	ProtoDir   string `toml:"proto_dir"`
	Reflection bool   `toml:"reflection"`

	ResolvedAddr string `toml:"-"`
}

type Log struct {
	Files []string `toml:"files"`
}

type Commands map[string]string

// Envs holds multi-environment configuration overrides.
// Each key is an environment name (e.g., "staging", "production"),
// and the value is a partial Config that overrides the base config.
type Envs map[string]EnvOverrides

// EnvOverrides represents per-environment configuration overrides.
// Only non-nil fields are applied on top of the base config.
type EnvOverrides struct {
	Database *DatabaseOverride `toml:"database,omitempty"`
	API      *APIOverride      `toml:"api,omitempty"`
	GRPC     *GRPCOverride     `toml:"grpc,omitempty"`
}

// DatabaseOverride contains optional overrides for the database section.
type DatabaseOverride struct {
	Source *string `toml:"source,omitempty"`
	Type   *string `toml:"type,omitempty"`
	Driver *string `toml:"driver,omitempty"`
}

// APIOverride contains optional overrides for the API section.
type APIOverride struct {
	BaseURL  *string           `toml:"base_url,omitempty"`
	Auth     *string           `toml:"auth,omitempty"`
	AuthType *string           `toml:"auth_type,omitempty"`
	Headers  map[string]string `toml:"headers,omitempty"`
}

// GRPCOverride contains optional overrides for the gRPC section.
type GRPCOverride struct {
	Addr       *string `toml:"addr,omitempty"`
	ProtoDir   *string `toml:"proto_dir,omitempty"`
	Reflection *bool   `toml:"reflection,omitempty"`
}

// ApplyEnv applies environment overrides to the base config.
// Only non-nil fields in the override are applied; zero-value fields are left unchanged.
func (c *Config) ApplyEnv(envName string) {
	if c.Envs == nil {
		return
	}
	override, ok := c.Envs[envName]
	if !ok {
		return
	}

	if override.Database != nil && c.Database != nil {
		if override.Database.Source != nil {
			c.Database.Source = *override.Database.Source
		}
		if override.Database.Type != nil {
			c.Database.Type = *override.Database.Type
		}
		if override.Database.Driver != nil {
			c.Database.Driver = *override.Database.Driver
		}
	}

	if override.API != nil && c.API != nil {
		if override.API.BaseURL != nil {
			c.API.BaseURL = *override.API.BaseURL
		}
		if override.API.Auth != nil {
			c.API.Auth = *override.API.Auth
		}
		if override.API.AuthType != nil {
			c.API.AuthType = *override.API.AuthType
		}
		if len(override.API.Headers) > 0 {
			if c.API.Headers == nil {
				c.API.Headers = make(map[string]string)
			}
			for k, v := range override.API.Headers {
				c.API.Headers[k] = v
			}
		}
	}

	if override.GRPC != nil && c.GRPC != nil {
		if override.GRPC.Addr != nil {
			c.GRPC.Addr = *override.GRPC.Addr
		}
		if override.GRPC.ProtoDir != nil {
			c.GRPC.ProtoDir = *override.GRPC.ProtoDir
		}
		if override.GRPC.Reflection != nil {
			c.GRPC.Reflection = *override.GRPC.Reflection
		}
	}
}

// ListEnvs returns the list of available environment names.
func (c *Config) ListEnvs() []string {
	if c.Envs == nil {
		return nil
	}
	names := make([]string, 0, len(c.Envs))
	for name := range c.Envs {
		names = append(names, name)
	}
	return names
}

// SetEnvVars stores the parsed .env variables on the config.
func (c *Config) SetEnvVars(env map[string]string) {
	c.envVars = env
}

// GetEnvVars returns the stored .env variables.
func (c *Config) GetEnvVars() map[string]string {
	if c.envVars == nil {
		return make(map[string]string)
	}
	return c.envVars
}
