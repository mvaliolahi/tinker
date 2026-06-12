package main

import (
        "fmt"
        "os"

        "github.com/mvaliolahi/tinker/internal/config"
        "github.com/mvaliolahi/tinker/internal/deps"
        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

// Version is set at build time via -ldflags: -X main.version=$(git describe --tags)
var version = "0.28.0"

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
                envCmd(), commandsCmd(), dockerCmd(), pluginCmd(),
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
