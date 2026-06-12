package contract

import (
        "fmt"
        "os"
        "path/filepath"
        "strings"

        "github.com/mvaliolahi/tinker/internal/detect"
)

func Generate(result *detect.Result, dir string) error {
        var sb strings.Builder

        sb.WriteString("# Tinker Configuration — edit to match your project\n")
        sb.WriteString("# https://github.com/mvaliolahi/tinker\n\n")

        writeDatabase(&sb, result.Database)
        writeAPI(&sb, result.API)
        writeGRPC(&sb, result.GRPC)
        writeLog(&sb, result.Log)
        writeCommands(&sb)
        writeDocker(&sb, result.Docker)
        writeEnvTemplate(&sb, result.Docker)

        return os.WriteFile(filepath.Join(dir, "tinker.toml"), []byte(sb.String()), 0644)
}

func writeDatabase(sb *strings.Builder, d *detect.DatabaseResult) {
        if d == nil {
                return
        }
        sb.WriteString("[database]\n")
        fmt.Fprintf(sb, "source = %q\n", d.Source)
        fmt.Fprintf(sb, "type = %q\n", d.Type)
        if d.MigrateDir != "" {
                fmt.Fprintf(sb, "migrate_dir = %q\n", d.MigrateDir)
        }
        if d.SeedDir != "" {
                fmt.Fprintf(sb, "seed_dir = %q\n", d.SeedDir)
        }
        sb.WriteString("\n")
}

func writeAPI(sb *strings.Builder, a *detect.APIResult) {
        if a == nil {
                return
        }
        sb.WriteString("[api]\n")
        if a.BaseURL != "" {
                fmt.Fprintf(sb, "base_url = %q\n", a.BaseURL)
        }
        if a.Spec != "" {
                fmt.Fprintf(sb, "spec = %q\n", a.Spec)
        }
        if a.Auth != "" {
                fmt.Fprintf(sb, "auth = %q\n", a.Auth)
        }
        if a.AuthType != "" {
                fmt.Fprintf(sb, "auth_type = %q\n", a.AuthType)
        }
        sb.WriteString("\n")
}

func writeGRPC(sb *strings.Builder, g *detect.GRPCResult) {
        if g == nil {
                return
        }
        sb.WriteString("[grpc]\n")
        if g.Addr != "" {
                fmt.Fprintf(sb, "addr = %q\n", g.Addr)
        }
        if g.ProtoDir != "" {
                fmt.Fprintf(sb, "proto_dir = %q\n", g.ProtoDir)
        }
        if g.Reflection {
                sb.WriteString("reflection = true\n")
        }
        sb.WriteString("\n")
}

func writeLog(sb *strings.Builder, l *detect.LogResult) {
        if l == nil || len(l.Files) == 0 {
                return
        }
        sb.WriteString("[log]\n")
        sb.WriteString("files = [\n")
        for _, f := range l.Files {
                fmt.Fprintf(sb, "  %q,\n", f)
        }
        sb.WriteString("]\n\n")
}

func writeCommands(sb *strings.Builder) {
        sb.WriteString("# [commands]\n")
        sb.WriteString("# migrate = \"go run ./cmd/migrate\"\n")
        sb.WriteString("# seed = \"go run ./cmd/seed\"\n\n")
}

func writeDocker(sb *strings.Builder, d *detect.DockerResult) {
        if d == nil {
                return
        }
        sb.WriteString("# Docker Compose detected: " + d.ComposeFile + "\n")
        sb.WriteString("# Use: tinker docker list\n\n")
}

func writeEnvTemplate(sb *strings.Builder, d *detect.DockerResult) {
        sb.WriteString("# Multi-environment overrides\n")
        sb.WriteString("# Use: tinker --env staging db\n")
        sb.WriteString("# [envs.staging.database]\n")
        sb.WriteString("# source = \"env:STAGING_DATABASE_URL\"\n")
        sb.WriteString("# [envs.staging.api]\n")
        sb.WriteString("# base_url = \"env:STAGING_API_BASE_URL\"\n")
        sb.WriteString("# auth = \"env:STAGING_API_TOKEN\"\n")
        sb.WriteString("# [envs.production.database]\n")
        sb.WriteString("# source = \"env:PRODUCTION_DATABASE_URL\"\n")
        sb.WriteString("# [envs.production.api]\n")
        sb.WriteString("# base_url = \"env:PRODUCTION_API_BASE_URL\"\n")
}

// SaveLogConfig updates or appends the [log] section in tinker.toml.
// If a [log] section already exists, it is replaced. Other content is preserved.
func SaveLogConfig(files []string, dir string) error {
        tomlPath := filepath.Join(dir, "tinker.toml")
        data, err := os.ReadFile(tomlPath)
        if err != nil {
                return fmt.Errorf("reading tinker.toml: %w", err)
        }

        content := string(data)

        // Build the new [log] section
        var logSection strings.Builder
        logSection.WriteString("[log]\n")
        logSection.WriteString("files = [\n")
        for _, f := range files {
                fmt.Fprintf(&logSection, "  %q,\n", f)
        }
        logSection.WriteString("]\n")

        // Check if [log] section already exists
        logStart := findSection(content, "log")

        if logStart >= 0 {
                // Find where this section ends (next section or EOF)
                logEnd := findNextSection(content, logStart)
                updated := content[:logStart] + logSection.String() + content[logEnd:]
                return os.WriteFile(tomlPath, []byte(updated), 0644)
        }

        // No [log] section — append before [commands] if it exists, otherwise at end
        commandsStart := findSection(content, "commands")
        if commandsStart >= 0 {
                updated := content[:commandsStart] + logSection.String() + "\n" + content[commandsStart:]
                return os.WriteFile(tomlPath, []byte(updated), 0644)
        }

        // Append at end
        updated := strings.TrimRight(content, "\n") + "\n\n" + logSection.String()
        return os.WriteFile(tomlPath, []byte(updated), 0644)
}

// findSection returns the byte index where [name] section starts, or -1.
func findSection(content, name string) int {
        // Match [name] at start of line (not [[name]] which is a table array)
        target := "[" + name + "]"
        for {
                idx := strings.Index(content, target)
                if idx < 0 {
                        return -1
                }
                // Make sure it's not a table array [[...]] by checking surrounding chars
                if idx > 0 && content[idx-1] == '[' {
                        content = content[idx+1:]
                        continue
                }
                // Check it's at start of line
                if idx == 0 || content[idx-1] == '\n' {
                        return idx
                }
                content = content[idx+1:]
        }
}

// findNextSection returns the byte index where the next section starts after the given position.
func findNextSection(content string, after int) int {
        // Skip past the section header line
        idx := after
        for idx < len(content) && content[idx] != '\n' {
                idx++
        }
        // Now scan for next [section] at start of line
        for idx < len(content) {
                lineStart := idx
                if content[idx] == '\n' {
                        idx++
                        lineStart = idx
                } else {
                        idx++
                        continue
                }
                if idx < len(content) && content[idx] == '[' && (idx+1 >= len(content) || content[idx+1] != '[') {
                        return lineStart
                }
        }
        return len(content)
}
