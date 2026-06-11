package main

import (
	"fmt"

	"github.com/mvaliolahi/tinker/internal/deps"
	"github.com/mvaliolahi/tinker/internal/ui"
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
			fmt.Println(ui.Header("  Dependencies"))
			fmt.Println()
			for _, dep := range deps.All() {
				fmt.Println(ui.DepStatus(dep.Name, deps.IsInstalled(dep.Name), dep.Purpose))
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
			fmt.Println(ui.Section("Installing Dependencies"))
			fmt.Println()
			failed := deps.InstallAll()
			if len(failed) > 0 {
				fmt.Println()
				fmt.Println(ui.Warning("Failed: " + deps.FormatList(failed)))
				fmt.Println(ui.Dim("  Try: GOPROXY=direct go install <module>@latest"))
			} else {
				fmt.Println()
				fmt.Println(ui.Success("All dependencies installed."))
			}
			return nil
		},
	}
}
