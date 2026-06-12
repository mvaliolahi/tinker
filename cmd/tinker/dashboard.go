package main

import (
	"fmt"
	"strings"

	"github.com/mvaliolahi/tinker/internal/config"
	"github.com/mvaliolahi/tinker/internal/detect"
)

// dbInfo extracts a short description from the database config, e.g. "SQLite • apartment.db"
func dbInfo(cfg *config.Config) string {
	if cfg.Database == nil {
		return ""
	}
	t := titleCase(cfg.Database.Type)
	url := cfg.Database.URL

	// Extract a short identifier from the URL
	switch cfg.Database.Type {
	case "sqlite3":
		// file:/path/to/db.sqlite3 or just /path/to/db.sqlite3
		path := strings.TrimPrefix(url, "file:")
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			return t + " • " + parts[len(parts)-1]
		}
	case "postgres", "postgresql":
		// postgres://user:pass@host:5432/dbname
		if strings.Contains(url, "@") && strings.Contains(url, "/") {
			afterAt := url[strings.Index(url, "@")+1:]
			if idx := strings.Index(afterAt, "?"); idx != -1 {
				afterAt = afterAt[:idx]
			}
			return t + " • " + afterAt
		}
	case "mysql":
		if strings.Contains(url, "@") && strings.Contains(url, "/") {
			afterAt := url[strings.Index(url, "@")+1:]
			if idx := strings.Index(afterAt, "?"); idx != -1 {
				afterAt = afterAt[:idx]
			}
			return t + " • " + afterAt
		}
	}
	return t
}

// apiInfo extracts a short description from the API config.
func apiInfo(cfg *config.Config) string {
	if cfg.API == nil {
		return ""
	}
	url := cfg.API.ResolvedBaseURL
	// Strip scheme
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	// Trim trailing slash
	url = strings.TrimRight(url, "/")
	return url
}

// grpcInfo extracts a short description from the gRPC config.
func grpcInfo(cfg *config.Config) string {
	if cfg.GRPC == nil {
		return ""
	}
	return cfg.GRPC.ResolvedAddr
}

// dockerDashboardInfo detects Docker Compose and returns (hasDocker, description).
func dockerDashboardInfo(root string) (bool, string) {
	result := detect.New(root).Detect()
	if result.Docker == nil {
		return false, ""
	}
	info := fmt.Sprintf("%d services (%s)", len(result.Docker.Services), result.Docker.ComposeFile)
	return true, info
}

// titleCase capitalizes the first letter of a string (replaces deprecated strings.Title).
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
