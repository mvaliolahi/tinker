package detect

import (
	"os"
	"path/filepath"
	"strings"
)

var envFiles = []string{".env", ".env.example", ".env.local", ".env.sample"}

func ParseEnvFiles(dir string) map[string]string {
	env := make(map[string]string)

	for _, name := range envFiles {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
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

			if v != "" {
				env[k] = v
			}
		}
	}

	return env
}
