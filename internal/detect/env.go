package detect

import (
	"os"
	"path/filepath"
	"strings"
)

func ParseEnvFile(dir string) map[string]string {
	env := make(map[string]string)
	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		return env
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		env[k] = v
	}

	return env
}
