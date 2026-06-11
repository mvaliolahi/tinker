package main

import (
	"fmt"
	"os"

	"github.com/mvaliolahi/tinker/internal/config"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

var projectDir string

func main() {
	root := &cobra.Command{
		Use:           "tinker",
		Short:         "Project-aware CLI for database, API, and gRPC interaction",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&projectDir, "project", "p", "", "project directory")

	root.AddCommand(initCmd(), dbCmd(), apiCmd(), grpcCmd(), runCmd(), makeCmd(), updateCmd(), depsCmd(), versionCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error(err.Error()))
		os.Exit(1)
	}
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
			fmt.Println(ui.Banner("0.7.0"))
		},
	}
}
