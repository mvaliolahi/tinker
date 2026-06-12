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
			fmt.Println()
			fmt.Println("  " + ui.Bold("Dependencies"))
			fmt.Println()

			headers := []string{"Tool", "Status", "Purpose"}
			var rows [][]string
			for _, dep := range deps.All() {
				status := "✓ installed"
				if !deps.IsInstalled(dep.Name) {
					status = "✗ missing"
				}
				rows = append(rows, []string{dep.Name, status, dep.Purpose})
			}
			fmt.Print(ui.Table(headers, rows))

			missing := deps.Check()
			if len(missing) > 0 {
				fmt.Println(ui.Hint("tinker deps install  to install missing tools"))
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
			fmt.Println()
			fmt.Println("  " + ui.Bold("Installing Dependencies"))
			fmt.Println()
			failed := deps.InstallAll()
			if len(failed) > 0 {
				fmt.Println()
				fmt.Println(ui.Warning("Failed: " + deps.FormatList(failed)))
				fmt.Println(ui.Hint("tinker deps install  to retry"))
			} else {
				fmt.Println()
				fmt.Println(ui.Success("All dependencies installed."))
			}
			return nil
		},
	}
}
