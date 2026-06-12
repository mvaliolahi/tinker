package main

import (
        "fmt"
        "os"
        "path/filepath"
        "strings"

        "github.com/mvaliolahi/tinker/internal/config"
        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
        cmd := &cobra.Command{
                Use:   "config",
                Short: "Manage tinker configuration",
        }

        cmd.AddCommand(configValidateCmd(), configShowCmd())

        return cmd
}

func configValidateCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "validate",
                Short: "Validate tinker.toml configuration",
                RunE: func(_ *cobra.Command, _ []string) error {
                        dir, err := resolveDir()
                        if err != nil {
                                return err
                        }

                        tomlPath := filepath.Join(dir, "tinker.toml")
                        if _, err := os.Stat(tomlPath); err != nil {
                                fmt.Println(ui.Error("tinker.toml not found"))
                                fmt.Println(ui.Hint("tinker init  to generate configuration"))
                                return nil
                        }

                        fmt.Println("  " + ui.CfgLabel() + " " + ui.Bold("Validate"))
                        fmt.Println(ui.KeyValue("file", tomlPath))
                        fmt.Println()

                        cfg, err := config.Load(dir)
                        if err != nil {
                                fmt.Println(ui.Error("Parse error: " + err.Error()))
                                return nil
                        }

                        issues := cfg.Validate()
                        if len(issues) == 0 {
                                fmt.Println(ui.Success("Configuration is valid"))
                                fmt.Println()

                                if cfg.Database != nil {
                                        fmt.Println(ui.Bullet("database", cfg.Database.Type+" ("+cfg.Database.Source+")"))
                                }
                                if cfg.API != nil {
                                        fmt.Println(ui.Bullet("api", cfg.API.BaseURL))
                                }
                                if cfg.GRPC != nil {
                                        fmt.Println(ui.Bullet("grpc", cfg.GRPC.Addr))
                                }
                                if cfg.Log != nil {
                                        fmt.Println(ui.Bullet("log", fmt.Sprintf("%d file(s)", len(cfg.Log.Files))))
                                }
                                return nil
                        }

                        fmt.Println(ui.Warning("Found issues:"))
                        fmt.Println()
                        for _, issue := range issues {
                                fmt.Println("  " + ui.Error(issue))
                        }
                        fmt.Println()
                        fmt.Println(ui.Hint("Edit tinker.toml to fix the issues above"))
                        return nil
                },
        }
}

func configShowCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "show",
                Short: "Display resolved configuration",
                RunE: func(_ *cobra.Command, _ []string) error {
                        cfg, root, err := loadConfig()
                        if err != nil {
                                return err
                        }

                        tomlPath := filepath.Join(root, "tinker.toml")
                        fmt.Println("  " + ui.CfgLabel() + " " + ui.Bold("Configuration"))
                        fmt.Println(ui.KeyValue("file", tomlPath))
                        fmt.Println()

                        if cfg.Database != nil {
                                fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Database"))
                                fmt.Println(ui.KeyValue("type", cfg.Database.Type))
                                fmt.Println(ui.KeyValue("source", cfg.Database.Source))
                                if cfg.Database.URL != "" {
                                        fmt.Println(ui.KeyValue("url", maskSecret(cfg.Database.URL)))
                                }
                                if cfg.Database.Driver != "" {
                                        fmt.Println(ui.KeyValue("driver", cfg.Database.Driver))
                                }
                                fmt.Println()
                        }

                        if cfg.API != nil {
                                fmt.Println("  " + ui.APILabel() + " " + ui.Bold("HTTP API"))
                                fmt.Println(ui.KeyValue("base_url", cfg.API.BaseURL))
                                if cfg.API.ResolvedBaseURL != "" {
                                        fmt.Println(ui.KeyValue("resolved", cfg.API.ResolvedBaseURL))
                                }
                                if cfg.API.Spec != "" {
                                        fmt.Println(ui.KeyValue("spec", cfg.API.Spec))
                                }
                                if cfg.API.Auth != "" {
                                        fmt.Println(ui.KeyValue("auth", cfg.API.Auth))
                                        fmt.Println(ui.KeyValue("auth_type", cfg.API.AuthType))
                                }
                                fmt.Println()
                        }

                        if cfg.GRPC != nil {
                                fmt.Println("  " + ui.GRPCLabel() + " " + ui.Bold("gRPC"))
                                fmt.Println(ui.KeyValue("addr", cfg.GRPC.Addr))
                                if cfg.GRPC.ResolvedAddr != "" {
                                        fmt.Println(ui.KeyValue("resolved", cfg.GRPC.ResolvedAddr))
                                }
                                fmt.Println(ui.KeyValue("proto_dir", cfg.GRPC.ProtoDir))
                                fmt.Println(ui.KeyValue("reflection", fmt.Sprintf("%v", cfg.GRPC.Reflection)))
                                fmt.Println()
                        }

                        if cfg.Log != nil {
                                fmt.Println("  " + ui.LogLabel() + " " + ui.Bold("Logs"))
                                for i, f := range cfg.Log.Files {
                                        fmt.Println(ui.KeyValue(fmt.Sprintf("file[%d]", i), f))
                                }
                                fmt.Println()
                        }

                        if len(cfg.Commands) > 0 {
                                fmt.Println("  " + ui.MakeLabel() + " " + ui.Bold("Commands"))
                                for name, cmd := range cfg.Commands {
                                        fmt.Println(ui.KeyValue(name, cmd))
                                }
                                fmt.Println()
                        }

                        if len(cfg.ListEnvs()) > 0 {
                                fmt.Println("  " + ui.CfgLabel() + " " + ui.Bold("Environments"))
                                for _, name := range cfg.ListEnvs() {
                                        override := cfg.Envs[name]
                                        parts := []string{}
                                        if override.Database != nil {
                                                parts = append(parts, "database")
                                        }
                                        if override.API != nil {
                                                parts = append(parts, "api")
                                        }
                                        if override.GRPC != nil {
                                                parts = append(parts, "grpc")
                                        }
                                        fmt.Println(ui.Bullet(name, fmt.Sprintf("overrides: %s", strings.Join(parts, ", "))))
                                }
                                fmt.Println()
                                fmt.Println(ui.Hint("Use: tinker --env <name> db"))
                                fmt.Println()
                        }

                        return nil
                },
        }
}

// maskSecret hides passwords in connection strings for safe display.
func maskSecret(s string) string {
        // Mask password in URL-style DSNs: scheme://user:pass@host/db → scheme://user:***@host/db
        if atIdx := strings.Index(s, "@"); atIdx != -1 {
                if colonIdx := strings.Index(s[:atIdx], ":"); colonIdx != -1 {
                        return s[:colonIdx+1] + "***" + s[atIdx:]
                }
        }
        // For file: paths, just return as-is (no secrets)
        return s
}
