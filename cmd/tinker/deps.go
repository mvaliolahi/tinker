package main

import (
	"fmt"

	"github.com/mvaliolahi/tinker/internal/deps"
	"github.com/spf13/cobra"
)

func depsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Check and install external dependencies",
	}

	cmd.AddCommand(depsListCmd(), depsInstallCmd())

	return cmd
}

func depsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List dependency status",
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, dep := range deps.All() {
				status := "✗ missing"
				if deps.IsInstalled(dep.Name) {
					status = "✓ installed"
				}
				fmt.Printf("  %-10s %s  (%s)\n", dep.Name, status, dep.Purpose)
			}
			return nil
		},
	}
}

func depsInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install missing dependencies",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Installing dependencies...")
			failed := deps.InstallAll()
			if len(failed) > 0 {
				fmt.Printf("\nFailed: %s — try with GOPROXY=direct:\n", deps.FormatList(failed))
				fmt.Println("  GOPROXY=direct go install <module>@latest")
			}
			return nil
		},
	}
}
