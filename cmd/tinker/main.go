package main

import (
        "fmt"
        "os"
        "strings"

        "github.com/mvaliolahi/tinker/internal/config"
        "github.com/mvaliolahi/tinker/internal/deps"
        "github.com/mvaliolahi/tinker/internal/detect"
        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

// Version is set at build time via -ldflags: -X main.version=$(git describe --tags)
var version = "0.25.4"

var projectDir string
var envName string

func main() {
        root := &cobra.Command{
                Use:           "tinker",
                Short:         "Project-aware CLI for database, API, and gRPC interaction",
                SilenceUsage:  true,
                SilenceErrors: true,
                RunE:          runDashboard,
        }

        root.PersistentFlags().StringVarP(&projectDir, "project", "p", "", "project directory")
        root.PersistentFlags().StringVarP(&envName, "env", "e", "", "environment (e.g., staging, production)")

        root.AddCommand(
                initCmd(), dbCmd(), apiCmd(), grpcCmd(), logCmd(),
                runCmd(), makeCmd(), updateCmd(), depsCmd(),
                configCmd(), versionCmd(), completionCmd(root),
                envCmd(), commandsCmd(), dockerCmd(),
        )

        if err := root.Execute(); err != nil {
                fmt.Fprintln(os.Stderr, ui.Error(err.Error()))
                os.Exit(1)
        }
}

func runDashboard(_ *cobra.Command, _ []string) error {
        cfg, root, err := loadConfig()
        if err != nil {
                fmt.Println(ui.Banner(version))
                fmt.Println()
                fmt.Println(ui.Warning("No tinker.toml found."))
                fmt.Println(ui.Hint("tinker init  to scan your project and generate config"))
                return nil
        }

        // Detect Docker Compose info for dashboard
        hasDocker, dockerInfo := dockerDashboardInfo(root)

        fmt.Print(ui.Dashboard(ui.DashboardConfig{
                ProjectDir:  root,
                HasDB:       cfg.Database != nil,
                DBInfo:      dbInfo(cfg),
                HasAPI:      cfg.API != nil,
                APIInfo:     apiInfo(cfg),
                HasGRPC:     cfg.GRPC != nil,
                GRPCInfo:    grpcInfo(cfg),
                HasLog:      cfg.Log != nil && len(cfg.Log.Files) > 0,
                HasDocker:   hasDocker,
                DockerInfo:  dockerInfo,
                MissingDeps: len(deps.Check()),
                Version:     version,
                Env:         envName,
        }))
        return nil
}

func loadConfig() (*config.Config, string, error) {
        dir := projectDir
        if dir == "" {
                var err error
                dir, err = os.Getwd()
                if err != nil {
                        return nil, "", err
                }
        }

        root, err := config.FindProjectRoot(dir)
        if err != nil {
                return nil, "", err
        }

        cfg, err := config.LoadWithEnv(root, envName)
        if err != nil {
                return nil, "", err
        }

        return cfg, root, nil
}

func resolveDir() (string, error) {
        dir := projectDir
        if dir == "" {
                var err error
                dir, err = os.Getwd()
                if err != nil {
                        return "", err
                }
        }
        return dir, nil
}

func versionCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "version",
                Short: "Print version",
                Run: func(_ *cobra.Command, _ []string) {
                        fmt.Println(ui.Banner(version))
                },
        }
}

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
