package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update tinker to the latest version",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Checking for updates...")

			if _, err := exec.LookPath("go"); err != nil {
				return fmt.Errorf("go not found — required for self-update")
			}

			cmd := exec.Command("go", "install", "github.com/mvaliolahi/tinker/cmd/tinker@latest")
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("update failed: %w", err)
			}

			fmt.Println("Updated successfully!")
			return versionCmd().Execute()
		},
	}
}
