package detect

import (
	"github.com/mvaliolahi/tinker/internal/env"
)

// ParseEnvFiles reads all .env files from a directory using the shared env parser.
func ParseEnvFiles(dir string) map[string]string {
	return env.ParseFiles(dir)
}
