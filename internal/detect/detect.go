package detect

type Result struct {
	Database *DatabaseResult `toml:"database,omitempty"`
	API      *APIResult      `toml:"api,omitempty"`
	GRPC     *GRPCResult     `toml:"grpc,omitempty"`
	Log      *LogResult      `toml:"log,omitempty"`
	Docker   *DockerResult   `toml:"docker,omitempty"`
}

type DatabaseResult struct {
	Source     string `toml:"source"`
	Type       string `toml:"type"`
	MigrateDir string `toml:"migrate_dir,omitempty"`
	SeedDir    string `toml:"seed_dir,omitempty"`
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
	return &Detector{dir: dir, env: ParseEnvFiles(dir)}
}

func (d *Detector) Detect() *Result {
	return &Result{
		Database: d.detectDatabase(),
		API:      d.detectAPI(),
		GRPC:     d.detectGRPC(),
		Log:      d.detectLog(),
		Docker:   d.detectDocker(),
	}
}

func (d *Detector) getEnv(key string) string {
	return d.env[key]
}
