package main

import (
	"fmt"

	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func envCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage multi-environment configuration",
	}

	cmd.AddCommand(envListCmd(), envShowCmd())

	return cmd
}

func envListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available environments",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			envs := cfg.ListEnvs()
			fmt.Println()
			fmt.Println("  " + ui.CfgLabel() + " " + ui.Bold("Environments"))
			fmt.Println()

			if len(envs) == 0 {
				fmt.Println(ui.Dim("  No environments configured"))
				fmt.Println()
				fmt.Println(ui.Hint("Add an [envs.staging] section to tinker.toml"))
				fmt.Println()
				fmt.Println(ui.Dim("  Example:"))
				fmt.Println(ui.Dim("    [envs.staging.database]"))
				fmt.Println(ui.Dim("      source = \"env:STAGING_DATABASE_URL\""))
				fmt.Println(ui.Dim("    [envs.staging.api]"))
				fmt.Println(ui.Dim("      base_url = \"env:STAGING_API_BASE_URL\""))
				return nil
			}

			for _, name := range envs {
				override := cfg.Envs[name]
				details := []string{}
				if override.Database != nil {
					details = append(details, "database")
				}
				if override.API != nil {
					details = append(details, "api")
				}
				if override.GRPC != nil {
					details = append(details, "grpc")
				}
				fmt.Println(ui.Bullet(name, fmt.Sprintf("overrides: %s", joinWithComma(details))))
			}
			fmt.Println()
			fmt.Println(ui.Hint("Use with: tinker --env " + envs[0] + " db"))
			return nil
		},
	}
}

func envShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [env-name]",
		Short: "Show configuration for a specific environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Load base config
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			envNameArg := args[0]
			override, ok := cfg.Envs[envNameArg]
			if !ok {
				fmt.Println(ui.Warning("Environment not found: " + envNameArg))
				fmt.Println(ui.Hint("Use 'tinker env list' to see available environments"))
				return nil
			}

			fmt.Println()
			fmt.Println("  " + ui.CfgLabel() + " " + ui.Bold("Environment: "+envNameArg))
			fmt.Println()

			if override.Database != nil {
				fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Database overrides"))
				if override.Database.Source != nil {
					fmt.Println(ui.KeyValue("source", *override.Database.Source))
				}
				if override.Database.Type != nil {
					fmt.Println(ui.KeyValue("type", *override.Database.Type))
				}
				if override.Database.Driver != nil {
					fmt.Println(ui.KeyValue("driver", *override.Database.Driver))
				}
				fmt.Println()
			}

			if override.API != nil {
				fmt.Println("  " + ui.APILabel() + " " + ui.Bold("API overrides"))
				if override.API.BaseURL != nil {
					fmt.Println(ui.KeyValue("base_url", *override.API.BaseURL))
				}
				if override.API.Auth != nil {
					fmt.Println(ui.KeyValue("auth", *override.API.Auth))
				}
				if override.API.AuthType != nil {
					fmt.Println(ui.KeyValue("auth_type", *override.API.AuthType))
				}
				fmt.Println()
			}

			if override.GRPC != nil {
				fmt.Println("  " + ui.GRPCLabel() + " " + ui.Bold("gRPC overrides"))
				if override.GRPC.Addr != nil {
					fmt.Println(ui.KeyValue("addr", *override.GRPC.Addr))
				}
				if override.GRPC.ProtoDir != nil {
					fmt.Println(ui.KeyValue("proto_dir", *override.GRPC.ProtoDir))
				}
				if override.GRPC.Reflection != nil {
					fmt.Println(ui.KeyValue("reflection", fmt.Sprintf("%v", *override.GRPC.Reflection)))
				}
				fmt.Println()
			}

			fmt.Println(ui.Hint("Use: tinker --env " + envNameArg + " db"))
			return nil
		},
	}
}

func joinWithComma(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}
