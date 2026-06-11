// Package detect handles automatic detection of project configuration
// for generating tinker.toml files.
package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectionResult holds what was found during project scanning.
type DetectionResult struct {
	Database *DatabaseDetection `toml:"database,omitempty"`
	API      *APIDetection      `toml:"api,omitempty"`
	GRPC     *GRPCDetection     `toml:"grpc,omitempty"`
}

// DatabaseDetection holds detected database configuration.
type DatabaseDetection struct {
	Source string `toml:"source"`
	Type   string `toml:"type"`
}

// APIDetection holds detected API configuration.
type APIDetection struct {
	BaseURL  string `toml:"base_url,omitempty"`
	Spec     string `toml:"spec,omitempty"`
	Auth     string `toml:"auth,omitempty"`
	AuthType string `toml:"auth_type,omitempty"`
}

// GRPCDetection holds detected gRPC configuration.
type GRPCDetection struct {
	Addr       string `toml:"addr,omitempty"`
	ProtoDir   string `toml:"proto_dir,omitempty"`
	Reflection bool   `toml:"reflection"`
}

// Detector scans a project directory and detects configuration.
type Detector struct {
	projectDir string
	envVars    map[string]string
}

// NewDetector creates a new Detector for the given project directory.
func NewDetector(projectDir string) *Detector {
	return &Detector{
		projectDir: projectDir,
		envVars:    make(map[string]string),
	}
}

// Detect scans the project and returns detected configuration.
func (d *Detector) Detect() (*DetectionResult, error) {
	result := &DetectionResult{}

	// Load .env if exists
	d.loadEnvFile()

	// Detect database
	db := d.detectDatabase()
	if db != nil {
		result.Database = db
	}

	// Detect API
	api := d.detectAPI()
	if api != nil {
		result.API = api
	}

	// Detect gRPC
	grpc := d.detectGRPC()
	if grpc != nil {
		result.GRPC = grpc
	}

	return result, nil
}

// loadEnvFile reads .env into d.envVars (without setting actual env vars).
func (d *Detector) loadEnvFile() {
	envPath := filepath.Join(d.projectDir, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			d.envVars[key] = val
		}
	}
}

// getEnv checks both actual environment and .env variables.
func (d *Detector) getEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return d.envVars[key]
}

// detectDatabase looks for common database configuration patterns.
func (d *Detector) detectDatabase() *DatabaseDetection {
	// Check common env var names for database URLs
	dbEnvNames := []string{
		"DATABASE_URL",
		"DB_URL",
		"DB_HOST",
		"POSTGRES_URL",
		"MYSQL_URL",
		"MONGO_URL",
		"MONGODB_URI",
		"SQLITE_PATH",
	}

	for _, envName := range dbEnvNames {
		if val := d.getEnv(envName); val != "" {
			dbType := detectDBType(val)
			return &DatabaseDetection{
				Source: fmt.Sprintf("env:%s", envName),
				Type:   dbType,
			}
		}
	}

	// Check for common DB file patterns
	dbFiles := []string{"db.sqlite3", "db.sqlite", "database.sqlite3", "test.db"}
	for _, f := range dbFiles {
		if _, err := os.Stat(filepath.Join(d.projectDir, f)); err == nil {
			return &DatabaseDetection{
				Source: f,
				Type:   "sqlite3",
			}
		}
	}

	return nil
}

// detectDBType infers the database type from a connection string.
func detectDBType(dsn string) string {
	lower := strings.ToLower(dsn)
	switch {
	case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"):
		return "postgres"
	case strings.HasPrefix(lower, "mysql://"), strings.HasPrefix(lower, "mysql2://"):
		return "mysql"
	case strings.HasPrefix(lower, "mongodb://"), strings.HasPrefix(lower, "mongodb+srv://"):
		return "mongodb"
	case strings.HasSuffix(lower, ".sqlite3"), strings.HasSuffix(lower, ".sqlite"), strings.HasSuffix(lower, ".db"):
		return "sqlite3"
	case strings.Contains(lower, "sqlserver://"), strings.Contains(lower, "mssql://"):
		return "sqlserver"
	default:
		// Try to infer from host patterns
		if strings.Contains(lower, "5432") {
			return "postgres"
		}
		if strings.Contains(lower, "3306") {
			return "mysql"
		}
		return "unknown"
	}
}

// detectAPI looks for common API configuration patterns.
func (d *Detector) detectAPI() *APIDetection {
	api := &APIDetection{}

	// Check for base URL
	urlEnvNames := []string{"API_BASE_URL", "API_URL", "BASE_URL", "SERVER_ADDR", "APP_URL"}
	for _, envName := range urlEnvNames {
		if d.getEnv(envName) != "" {
			api.BaseURL = fmt.Sprintf("env:%s", envName)
			break
		}
	}

	// Check for auth token
	authEnvNames := []string{"API_TOKEN", "API_KEY", "AUTH_TOKEN", "BEARER_TOKEN", "API_AUTH_TOKEN"}
	for _, envName := range authEnvNames {
		if d.getEnv(envName) != "" {
			api.Auth = fmt.Sprintf("env:%s", envName)
			api.AuthType = "bearer"
			break
		}
	}

	// Check for OpenAPI spec files
	specFiles := []string{
		"openapi.yaml", "openapi.yml", "openapi.json",
		"swagger.yaml", "swagger.yml", "swagger.json",
		"api/openapi.yaml", "api/openapi.yml", "api/openapi.json",
		"docs/openapi.yaml", "docs/swagger.json",
	}
	for _, f := range specFiles {
		if _, err := os.Stat(filepath.Join(d.projectDir, f)); err == nil {
			api.Spec = f
			break
		}
	}

	if api.BaseURL == "" && api.Spec == "" && api.Auth == "" {
		return nil
	}

	return api
}

// detectGRPC looks for gRPC configuration patterns.
func (d *Detector) detectGRPC() *GRPCDetection {
	grpc := &GRPCDetection{}

	// Check for gRPC address
	addrEnvNames := []string{"GRPC_ADDR", "GRPC_ADDRESS", "GRPC_HOST"}
	for _, envName := range addrEnvNames {
		if d.getEnv(envName) != "" {
			grpc.Addr = fmt.Sprintf("env:%s", envName)
			break
		}
	}

	// Check for proto files
	protoDirs := []string{"proto", "protos", "api/proto", "api/protos", "pkg/proto"}
	for _, dir := range protoDirs {
		fullPath := filepath.Join(d.projectDir, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			// Check if directory contains .proto files
			if hasProtoFiles(fullPath) {
				grpc.ProtoDir = dir
				break
			}
		}
	}

	if grpc.Addr == "" && grpc.ProtoDir == "" {
		return nil
	}

	// Default to reflection enabled if we found proto files or address
	grpc.Reflection = true

	return grpc
}

// hasProtoFiles checks if a directory contains .proto files.
func hasProtoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".proto") {
			return true
		}
	}
	return false
}
