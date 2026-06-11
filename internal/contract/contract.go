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
	writeCommands(&sb)

	return os.WriteFile(filepath.Join(dir, "tinker.toml"), []byte(sb.String()), 0644)
}

func writeDatabase(sb *strings.Builder, d *detect.DatabaseResult) {
	if d == nil {
		return
	}
	sb.WriteString("[database]\n")
	fmt.Fprintf(sb, "source = %q\n", d.Source)
	fmt.Fprintf(sb, "type = %q\n\n", d.Type)
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

func writeCommands(sb *strings.Builder) {
	sb.WriteString("# [commands]\n")
	sb.WriteString("# migrate = \"go run ./cmd/migrate\"\n")
	sb.WriteString("# seed = \"go run ./cmd/seed\"\n\n")
}
