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

			fmt.Println(ui.Banner(version))
			fmt.Println()
			fmt.Println(ui.Step(1, "Scanning project"))
			fmt.Println(ui.Dim("  " + dir))
			fmt.Println()

			for _, name := range []string{".env", ".env.example", ".env.local"} {
				if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
					fmt.Println(ui.Bullet("found", name))
				}
			}

			// Check for docker-compose
			for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
				if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
					fmt.Println(ui.Bullet("found", name))
				}
			}

			fmt.Println()
			fmt.Println(ui.Step(2, "Detecting services"))
			fmt.Println()

			result := detect.New(dir).Detect()
			printDetection(result)

			if result.Database == nil && result.API == nil && result.GRPC == nil && result.Log == nil && result.Docker == nil {
				fmt.Println()
				fmt.Println(ui.Warning("Nothing detected."))
				fmt.Println(ui.Hint("Create tinker.toml manually — see docs for examples."))
				return nil
			}

			if err := contract.Generate(result, dir); err != nil {
				return err
			}

			fmt.Println()
			fmt.Println(ui.StepDone(2, "Generated tinker.toml"))
			fmt.Println(ui.Dim("  Review and adjust as needed."))

			fmt.Println()
			fmt.Println(ui.Step(3, "Installing dependencies"))
			fmt.Println()

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

			fmt.Println()
			if failed > 0 {
				fmt.Println(ui.Warning(fmt.Sprintf("%d dependencies failed to install.", failed)))
				fmt.Println(ui.Hint("tinker deps install  to retry"))
			} else {
				fmt.Println(ui.StepDone(3, "All dependencies installed"))
			}

			fmt.Println()
			fmt.Println(ui.Step(4, "Ready!"))
			fmt.Println(ui.Hint("tinker db              Connect to your database"))
			fmt.Println(ui.Hint("tinker api GET /path   Call an API endpoint"))
			fmt.Println(ui.Hint("tinker api endpoints   List API endpoints from spec"))
			fmt.Println(ui.Hint("tinker api explore     Interactive API explorer"))
			fmt.Println(ui.Hint("tinker cmd <name>      Run a custom command"))
			fmt.Println(ui.Hint("tinker docker          Inspect Docker Compose services"))
			fmt.Println(ui.Hint("tinker --env staging   Switch environment"))
			fmt.Println(ui.Hint("tinker                 Show project dashboard"))
			return nil
		},
	}
}

func printDetection(r *detect.Result) {
	if r.Database != nil {
		fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Database"))
		fmt.Println(ui.KeyValue("type", r.Database.Type))
		fmt.Println(ui.KeyValue("source", r.Database.Source))
		fmt.Println()
	}
	if r.API != nil {
		fmt.Println("  " + ui.APILabel() + " " + ui.Bold("API"))
		fmt.Println(ui.KeyValue("base_url", r.API.BaseURL))
		if r.API.Spec != "" {
			fmt.Println(ui.KeyValue("spec", r.API.Spec))
		}
		fmt.Println()
	}
	if r.GRPC != nil {
		fmt.Println("  " + ui.GRPCLabel() + " " + ui.Bold("gRPC"))
		fmt.Println(ui.KeyValue("addr", r.GRPC.Addr))
		if r.GRPC.ProtoDir != "" {
			fmt.Println(ui.KeyValue("proto_dir", r.GRPC.ProtoDir))
		}
		fmt.Println()
	}
	if r.Log != nil {
		fmt.Println("  " + ui.LogLabel() + " " + ui.Bold("Logs"))
		for _, f := range r.Log.Files {
			fmt.Println(ui.Bullet("file", f))
		}
		fmt.Println()
	}
	if r.Docker != nil {
		fmt.Println("  " + ui.Bold("Docker Compose"))
		fmt.Println(ui.KeyValue("file", r.Docker.ComposeFile))
		for _, svc := range r.Docker.Services {
			label := svc.Name
			if svc.Image != "" {
				label += " (" + svc.Image + ")"
			}
			fmt.Println(ui.Bullet("service", label))
		}
		fmt.Println()
	}
}
