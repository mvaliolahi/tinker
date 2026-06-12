package detect

import (
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// DockerResult represents detected Docker service information.
type DockerResult struct {
	ComposeFile string              `toml:"compose_file,omitempty"`
	Services    []DockerServiceInfo `toml:"services,omitempty"`
}

// DockerServiceInfo represents a single Docker Compose service.
type DockerServiceInfo struct {
	Name    string `json:"name"`
	Image   string `json:"image,omitempty"`
	Ports   string `json:"ports,omitempty"`
	HasDB   bool   `json:"has_db"`
	HasAPI  bool   `json:"has_api"`
	HasGRPC bool   `json:"has_grpc"`
}

// composeFiles lists common docker-compose filenames to check.
var composeFiles = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
	// Environment-specific compose files
	"docker-compose.staging.yml",
	"docker-compose.prod.yml",
	"docker-compose.production.yml",
	"docker-compose.dev.yml",
	"docker-compose.development.yml",
	"docker-compose.local.yml",
	"docker-compose.override.yml",
}

// detectDocker scans for docker-compose files and extracts service information.
func (d *Detector) detectDocker() *DockerResult {
	for _, f := range composeFiles {
		fullPath := filepath.Join(d.dir, f)
		if _, err := os.Stat(fullPath); err != nil {
			continue
		}

		result := &DockerResult{ComposeFile: f}
		services := d.parseComposeFile(fullPath)
		if len(services) > 0 {
			result.Services = services
			return result
		}
	}
	return nil
}

// parseComposeFile reads a docker-compose file and extracts service information.
func (d *Detector) parseComposeFile(path string) []DockerServiceInfo {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var compose struct {
		Services map[string]struct {
			Image   string   `yaml:"image"`
			Ports   []string `yaml:"ports"`
			Command string   `yaml:"command"`
			Build   struct {
				Context string `yaml:"context"`
			} `yaml:"build"`
			Environment map[string]string `yaml:"environment"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil
	}

	var services []DockerServiceInfo
	for name, svc := range compose.Services {
		info := DockerServiceInfo{
			Name:  name,
			Image: svc.Image,
		}

		// Format ports
		if len(svc.Ports) > 0 {
			info.Ports = strings.Join(svc.Ports, ", ")
		}

		// Heuristic: detect service type by image name, command, or environment
		imageLower := strings.ToLower(svc.Image)
		nameLower := strings.ToLower(name)
		envKeys := make([]string, 0, len(svc.Environment))
		for k := range svc.Environment {
			envKeys = append(envKeys, strings.ToLower(k))
		}

		// Detect database services
		if containsAny(imageLower, "postgres", "mysql", "mariadb", "mongo", "redis", "sqlite", "cockroach") ||
			containsAny(nameLower, "db", "database", "postgres", "mysql", "mongo", "redis") ||
			containsAnyEnv(envKeys, "postgres", "mysql", "database_url", "db_host") {
			info.HasDB = true
		}

		// Detect API services
		if containsAny(nameLower, "api", "server", "app", "web", "backend", "rest") ||
			containsAnyEnv(envKeys, "api_base_url", "server_addr", "app_url", "port") {
			info.HasAPI = true
		}

		// Detect gRPC services
		if containsAny(nameLower, "grpc", "rpc") ||
			containsAnyEnv(envKeys, "grpc_addr", "grpc_port") {
			info.HasGRPC = true
		}

		services = append(services, info)
	}

	return services
}

// containsAny checks if s contains any of the given substrings.
func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// containsAnyEnv checks if any env key contains any of the given substrings.
func containsAnyEnv(envKeys []string, substrings ...string) bool {
	for _, key := range envKeys {
		for _, sub := range substrings {
			if strings.Contains(key, sub) {
				return true
			}
		}
	}
	return false
}
