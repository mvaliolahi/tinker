package main

import (
	"os"

	"github.com/mvaliolahi/tinker/internal/run"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [code]",
		Short: "Execute one-off Go code in project context",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			projectRoot := ""

			cfg, root, err := loadConfig()
			if err != nil {
				projectRoot, _ = os.Getwd()
			} else {
				_ = cfg
				projectRoot = root
			}

			code, imports := run.ParseCode(args[0])
			return run.NewRunner(run.Config{
				ProjectDir: projectRoot,
				Code:       code,
				Imports:    imports,
			}).Run()
		},
	}
}
