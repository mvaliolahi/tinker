package config

// env.go is kept for backward compatibility but now delegates to the shared env package.
// The LoadEnvFile function that mutated os.Setenv has been removed.
// Use internal/env.ParseFile() and internal/env.ParseFiles() instead.
