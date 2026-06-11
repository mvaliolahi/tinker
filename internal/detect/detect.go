package detect

import "os"

type Result struct {
	Database *DatabaseResult `toml:"database,omitempty"`
	API      *APIResult      `toml:"api,omitempty"`
	GRPC     *GRPCResult     `toml:"grpc,omitempty"`
}

type DatabaseResult struct {
	Source string `toml:"source"`
	Type   string `toml:"type"`
}

type APIResult struct {
	BaseURL  string `toml:"base_url,omitempty"`
	Spec     string `toml:"spec,omitempty"`
	Auth     string `toml:"auth,omitempty"`
	AuthType string `toml:"auth_type,omitempty"`
}

type GRPCResult struct {
	Addr       string `toml:"addr,omitempty"`
	ProtoDir   string `toml:"proto_dir,omitempty"`
	Reflection bool   `toml:"reflection"`
}

type Detector struct {
	dir string
	env map[string]string
}

func New(dir string) *Detector {
	return &Detector{dir: dir, env: ParseEnvFile(dir)}
}

func (d *Detector) Detect() *Result {
	return &Result{
		Database: d.detectDatabase(),
		API:      d.detectAPI(),
		GRPC:     d.detectGRPC(),
	}
}

func (d *Detector) getEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return d.env[key]
}
