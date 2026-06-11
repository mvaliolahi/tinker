package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mvaliolahi/tinker/internal/contract"
	"github.com/mvaliolahi/tinker/internal/detect"
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
				fmt.Fprintf(os.Stderr, "tinker.toml already exists — remove it to regenerate\n")
				return nil
			}

			fmt.Printf("Scanning %s...\n", dir)
			printEnvScan(dir)

			result := detect.New(dir).Detect()
			printDetection(result)

			if result.Database == nil && result.API == nil && result.GRPC == nil {
				fmt.Println("\nNothing detected. Create tinker.toml manually — see docs for examples.")
				return nil
			}

			if err := contract.Generate(result, dir); err != nil {
				return err
			}

			fmt.Printf("\nGenerated tinker.toml — review and adjust as needed.\n")
			return nil
		},
	}
}

func printEnvScan(dir string) {
	for _, name := range []string{".env", ".env.example", ".env.local"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			fmt.Printf("  found %s\n", name)
		}
	}
}

func printDetection(r *detect.Result) {
	if r.Database != nil {
		fmt.Printf("  [database] type=%s source=%s\n", r.Database.Type, r.Database.Source)
	}
	if r.API != nil {
		fmt.Printf("  [api] base_url=%s spec=%s\n", r.API.BaseURL, r.API.Spec)
	}
	if r.GRPC != nil {
		fmt.Printf("  [grpc] addr=%s proto_dir=%s\n", r.GRPC.Addr, r.GRPC.ProtoDir)
	}
}
