package config

type Config struct {
	Database *Database `toml:"database"`
	API      *API      `toml:"api"`
	GRPC     *GRPC     `toml:"grpc"`
	Log      *Log      `toml:"log"`
	Commands Commands  `toml:"commands"`

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
