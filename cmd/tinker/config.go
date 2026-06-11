package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mvaliolahi/tinker/internal/config"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage tinker configuration",
	}

	cmd.AddCommand(configValidateCmd())

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
