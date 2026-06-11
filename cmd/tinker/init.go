package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mvaliolahi/tinker/internal/contract"
	"github.com/mvaliolahi/tinker/internal/deps"
	"github.com/mvaliolahi/tinker/internal/detect"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Detect project config and generate tinker.toml",
		RunE: func(_ *cobra.Command, _ []string) error {
			dir, err := resolveDir()
			if err != nil {
				return err
			}

			if _, err := os.Stat(filepath.Join(dir, "tinker.toml")); err == nil {
				fmt.Fprintln(os.Stderr, ui.Warning("tinker.toml already exists — remove it to regenerate"))
				return nil
			}

			fmt.Println(ui.Banner("0.7.1"))
			fmt.Println()

			fmt.Println(ui.Header("  Scanning project..."))
			fmt.Println(ui.Dim("  " + dir))
			printEnvScan(dir)

			result := detect.New(dir).Detect()
			printDetection(result)

			if result.Database == nil && result.API == nil && result.GRPC == nil {
				fmt.Println()
				fmt.Println(ui.Warning("Nothing detected."))
				fmt.Println(ui.Dim("  Create tinker.toml manually — see docs for examples."))
				return nil
			}

			if err := contract.Generate(result, dir); err != nil {
				return err
			}

			fmt.Println()
			fmt.Println(ui.Success("Generated tinker.toml"))
			fmt.Println(ui.Dim("  Review and adjust as needed."))

			fmt.Println(ui.Section("Installing Dependencies"))
			var failed int
			if result.Database != nil {
				failed += len(deps.InstallForPurpose("database"))
			}
			if result.API != nil {
				failed += len(deps.InstallForPurpose("api"))
			}
			if result.GRPC != nil {
				failed += len(deps.InstallForPurpose("grpc"))
			}

			if failed > 0 {
				fmt.Println()
				fmt.Println(ui.Warning("Some dependencies failed to install."))
				fmt.Println(ui.Dim("  Run 'tinker deps install' to retry"))
				fmt.Println(ui.Dim("  Or: GOPROXY=direct go install <module>@latest"))
			} else {
				fmt.Println()
				fmt.Println(ui.Success("All dependencies installed."))
			}

			fmt.Println()
			fmt.Println(ui.Dim("  Run 'tinker db' to connect to your database."))
			return nil
		},
	}
}

func printEnvScan(dir string) {
	for _, name := range []string{".env", ".env.example", ".env.local"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			fmt.Println(ui.Bullet("found", name))
		}
	}
}

func printDetection(r *detect.Result) {
	if r.Database != nil {
		fmt.Println()
		fmt.Println("  " + ui.DBLabel() + " " + ui.Header("Database"))
		fmt.Println(ui.KeyValue("  type", r.Database.Type))
		fmt.Println(ui.KeyValue("  source", r.Database.Source))
	}
	if r.API != nil {
		fmt.Println()
		fmt.Println("  " + ui.APILabel() + " " + ui.Header("API"))
		fmt.Println(ui.KeyValue("  base_url", r.API.BaseURL))
		if r.API.Spec != "" {
			fmt.Println(ui.KeyValue("  spec", r.API.Spec))
		}
	}
	if r.GRPC != nil {
		fmt.Println()
		fmt.Println("  " + ui.GRPCLabel() + " " + ui.Header("gRPC"))
		fmt.Println(ui.KeyValue("  addr", r.GRPC.Addr))
		if r.GRPC.ProtoDir != "" {
			fmt.Println(ui.KeyValue("  proto_dir", r.GRPC.ProtoDir))
		}
	}
}
