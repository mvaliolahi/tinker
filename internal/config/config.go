package config

type Config struct {
	Database *Database `toml:"database"`
	API      *API      `toml:"api"`
	GRPC     *GRPC     `toml:"grpc"`
	Commands Commands  `toml:"commands"`
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

type Commands map[string]string
