package main

import (
        "fmt"
        "os"

        "github.com/mvaliolahi/tinker/internal/config"
        "github.com/mvaliolahi/tinker/internal/deps"
        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

const version = "0.17.0"

var projectDir string

func main() {
        root := &cobra.Command{
                Use:           "tinker",
                Short:         "Project-aware CLI for database, API, and gRPC interaction",
                SilenceUsage:  true,
                SilenceErrors: true,
                RunE:          runDashboard,
        }

        root.PersistentFlags().StringVarP(&projectDir, "project", "p", "", "project directory")

        root.AddCommand(initCmd(), dbCmd(), apiCmd(), grpcCmd(), logCmd(), runCmd(), makeCmd(), updateCmd(), depsCmd(), versionCmd())

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

        hasDB := cfg.Database != nil
        hasAPI := cfg.API != nil
        hasGRPC := cfg.GRPC != nil
        hasLog := cfg.Log != nil && len(cfg.Log.Files) > 0
        missing := len(deps.Check())

        fmt.Print(ui.Dashboard(root, root, hasDB, hasAPI, hasGRPC, hasLog, missing, version))
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

        cfg, err := config.Load(root)
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
