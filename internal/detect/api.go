package detect

import (
	"fmt"
	"os"
	"path/filepath"
)

var urlEnvNames = []string{"API_BASE_URL", "API_URL", "BASE_URL", "SERVER_ADDR", "APP_URL"}
var authEnvNames = []string{"API_TOKEN", "API_KEY", "AUTH_TOKEN", "BEARER_TOKEN", "API_AUTH_TOKEN"}

var specFiles = []string{
	"openapi.yaml", "openapi.yml", "openapi.json",
	"swagger.yaml", "swagger.yml", "swagger.json",
	"api/openapi.yaml", "api/openapi.yml", "docs/openapi.yaml", "docs/swagger.json",
}

func (d *Detector) detectAPI() *APIResult {
	r := &APIResult{}

	for _, name := range urlEnvNames {
		if d.getEnv(name) != "" {
			r.BaseURL = fmt.Sprintf("env:%s", name)
			break
		}
	}

	for _, name := range authEnvNames {
		if d.getEnv(name) != "" {
			r.Auth = fmt.Sprintf("env:%s", name)
			r.AuthType = "bearer"
			break
		}
	}

	for _, f := range specFiles {
		if _, err := os.Stat(filepath.Join(d.dir, f)); err == nil {
			r.Spec = f
			break
		}
	}

	if r.BaseURL == "" && r.Spec == "" && r.Auth == "" {
		return nil
	}
	return r
}
