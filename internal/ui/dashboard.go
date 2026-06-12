package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DashboardConfig holds all parameters for rendering the dashboard.
type DashboardConfig struct {
	ProjectDir  string
	HasDB       bool
	DBInfo      string // e.g. "SQLite • apartment.db" or "Postgres • localhost:5432/mydb"
	HasAPI      bool
	APIInfo     string // e.g. "localhost:8080"
	HasGRPC     bool
	GRPCInfo    string // e.g. "localhost:50051"
	HasLog      bool
	HasDocker   bool
	DockerInfo  string // e.g. "3 services (docker-compose.yml)"
	MissingDeps int
	Version     string
	Env         string // active environment name (empty = default)
}

// Dashboard renders the main dashboard view.
func Dashboard(cfg DashboardConfig) string {
	var sb strings.Builder

	sb.WriteString(Banner(cfg.Version))
	sb.WriteString("\n\n")

	sb.WriteString("  ")
	sb.WriteString(labelStyle.Render("Project: "))
	sb.WriteString(valueStyle.Render(cfgDir(cfg.ProjectDir)))
	if cfg.Env != "" {
		sb.WriteString("  ")
		sb.WriteString(labelStyle.Render("Env: "))
		sb.WriteString(accent.Bold(true).Render(cfg.Env))
	}
	sb.WriteString("\n\n")

	sb.WriteString("  ")
	sb.WriteString(labelStyle.Render("Services"))
	sb.WriteString("\n")

	sb.WriteString(dashboardServices(cfg))

	if !cfg.HasDB && !cfg.HasAPI && !cfg.HasGRPC && !cfg.HasLog && !cfg.HasDocker {
		sb.WriteString(Dim("  No services configured"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	if cfg.MissingDeps > 0 {
		sb.WriteString("  ")
		sb.WriteString(warning.Render(fmt.Sprintf("⚠ %d missing dependencies", cfg.MissingDeps)))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker deps install"))
		sb.WriteString("\n\n")
	}

	sb.WriteString(dashboardCommands(cfg))

	return sb.String()
}

// dashboardServices renders the services section of the dashboard.
func dashboardServices(cfg DashboardConfig) string {
	var sb strings.Builder

	if cfg.HasDB {
		sb.WriteString("  ")
		sb.WriteString(DBLabel())
		sb.WriteString(" ")
		sb.WriteString(success.Render("Database"))
		if cfg.DBInfo != "" {
			sb.WriteString(" ")
			sb.WriteString(dimStyle.Render("• " + cfg.DBInfo))
		}
		sb.WriteString("\n")
	}
	if cfg.HasAPI {
		sb.WriteString("  ")
		sb.WriteString(APILabel())
		sb.WriteString(" ")
		sb.WriteString(success.Render("HTTP API"))
		if cfg.APIInfo != "" {
			sb.WriteString(" ")
			sb.WriteString(dimStyle.Render("• " + cfg.APIInfo))
		}
		sb.WriteString("\n")
	}
	if cfg.HasGRPC {
		sb.WriteString("  ")
		sb.WriteString(GRPCLabel())
		sb.WriteString(" ")
		sb.WriteString(success.Render("gRPC"))
		if cfg.GRPCInfo != "" {
			sb.WriteString(" ")
			sb.WriteString(dimStyle.Render("• " + cfg.GRPCInfo))
		}
		sb.WriteString("\n")
	}
	if cfg.HasLog {
		sb.WriteString("  ")
		sb.WriteString(LogLabel())
		sb.WriteString(" ")
		sb.WriteString(success.Render("Logs"))
		sb.WriteString("\n")
	}
	if cfg.HasDocker {
		sb.WriteString("  ")
		sb.WriteString(badge.Foreground(lipgloss.Color("8")).Background(lipgloss.Color("0")).Render(" DOCKER "))
		sb.WriteString(" ")
		sb.WriteString(success.Render("Compose"))
		if cfg.DockerInfo != "" {
			sb.WriteString(" ")
			sb.WriteString(dimStyle.Render("• " + cfg.DockerInfo))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// dashboardCommands renders the quick commands section of the dashboard.
func dashboardCommands(cfg DashboardConfig) string {
	var sb strings.Builder

	sb.WriteString("  ")
	sb.WriteString(labelStyle.Render("Quick Commands"))
	sb.WriteString("\n")

	if cfg.HasDB {
		sb.WriteString(Hint("tinker db                Interactive database session"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db ls             List tables"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db desc users     Show table schema"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db idx users      Show table indexes"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db s users        Show CREATE TABLE"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db c users        Count rows"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db f users 1      Find row by ID"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db sql \"SELECT\"   Run SQL (native)"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db ping           Test connectivity"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db size           Table row counts"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db migrate up     Run pending migrations"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db seed           Seed the database"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker db explore        Interactive DB browser"))
		sb.WriteString("\n")
	}
	if cfg.HasAPI {
		sb.WriteString(Hint("tinker api GET /users  Call an API endpoint"))
		sb.WriteString("\n")
	}
	if cfg.HasGRPC {
		sb.WriteString(Hint("tinker grpc list       List gRPC services"))
		sb.WriteString("\n")
	}
	if cfg.HasLog {
		sb.WriteString(Hint("tinker log             Show log file"))
		sb.WriteString("\n")
		sb.WriteString(Hint("tinker log tail        Follow log in real-time"))
		sb.WriteString("\n")
	}
	sb.WriteString(Hint("tinker config validate Check configuration"))
	sb.WriteString("\n")
	sb.WriteString(Hint("tinker make build      Run a Makefile target"))
	sb.WriteString("\n")

	return sb.String()
}

func cfgDir(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return path
}
